package objects

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"strconv"
	"strings"

	. "raytracer/common"
	. "raytracer/material"
)

func NewMeshFromFile(path string, material Material, scale float64) Hittable_list {
	var triangles Hittable_list

	file, err := os.Open(path)
	if err != nil {
		fmt.Println("Error Loading Mesh from ", path)
		return triangles
	}
	defer file.Close()

	return NewMeshFromReader(file, material, scale)
}

func NewMeshFromReader(r io.Reader, material Material, scale float64) Hittable_list {
	var triangles Hittable_list

	// Use bufio.Reader to peek at the file content
	br := bufio.NewReader(r)

	// Peek first 5 bytes to check for "solid"
	header, err := br.Peek(5)
	if err == nil && string(header) == "solid" {
		// ASCII STL
		scanner := bufio.NewScanner(br)

		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)

			if len(fields) > 0 && fields[0] == "facet" {

				scanner.Scan()
				scanner.Scan()

				// read vertex 1 data
				line = scanner.Text()
				fields = strings.Fields(line)
				x, _ := strconv.ParseFloat(fields[1], 64)
				y, _ := strconv.ParseFloat(fields[2], 64)
				z, _ := strconv.ParseFloat(fields[3], 64)
				a := NewPoint3(x, y, z)

				scanner.Scan()
				// read vertex 2 data
				line = scanner.Text()
				fields = strings.Fields(line)
				x, _ = strconv.ParseFloat(fields[1], 64)
				y, _ = strconv.ParseFloat(fields[2], 64)
				z, _ = strconv.ParseFloat(fields[3], 64)
				b := NewPoint3(x, y, z)

				scanner.Scan()
				// read vertex 3 data
				line = scanner.Text()
				fields = strings.Fields(line)
				x, _ = strconv.ParseFloat(fields[1], 64)
				y, _ = strconv.ParseFloat(fields[2], 64)
				z, _ = strconv.ParseFloat(fields[3], 64)
				c := NewPoint3(x, y, z)

				tri := Triangle(a.Mult(scale), b.Mult(scale), c.Mult(scale), material)
				triangles.Add(&tri)
			}
		}
	} else {
		// Binary STL
		// Read 80 bytes header
		header := make([]byte, 80)
		if _, err := io.ReadFull(br, header); err != nil {
			return triangles
		}

		// Read number of triangles (uint32)
		var count uint32
		if err := binary.Read(br, binary.LittleEndian, &count); err != nil {
			return triangles
		}

		// Read triangles
		// Each triangle is 50 bytes: 12 bytes Normal, 12 bytes V1, 12 bytes V2, 12 bytes V3, 2 bytes Attribute
		for i := 0; i < int(count); i++ {
			var data [50]byte
			if _, err := io.ReadFull(br, data[:]); err != nil {
				break
			}

			// We skip the normal (bytes 0-11) and attribute (bytes 48-49)
			
			// Vertex 1
			v1x := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[12:16])))
			v1y := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[16:20])))
			v1z := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[20:24])))
			a := NewPoint3(v1x, v1y, v1z)

			// Vertex 2
			v2x := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[24:28])))
			v2y := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[28:32])))
			v2z := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[32:36])))
			b := NewPoint3(v2x, v2y, v2z)

			// Vertex 3
			v3x := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[36:40])))
			v3y := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[40:44])))
			v3z := float64(math.Float32frombits(binary.LittleEndian.Uint32(data[44:48])))
			c := NewPoint3(v3x, v3y, v3z)

			tri := Triangle(a.Mult(scale), b.Mult(scale), c.Mult(scale), material)
			triangles.Add(&tri)
		}
	}

	return triangles
}
