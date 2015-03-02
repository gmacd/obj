package obj

import (
	"strings"
	"testing"
)

func TestParseMaterialName(t *testing.T) {
	ml, err := parseMaterialName([]string{})
	if err == nil {
		t.Errorf("Expected error")
	}

	ml, err = parseMaterialName([]string{"foo"})
	if ml != "foo" {
		t.Errorf("Expected foo")
	}
}

func TestParseObjectName(t *testing.T) {
	ml, err := parseObjectName([]string{})
	if err == nil {
		t.Errorf("Expected error")
	}

	ml, err = parseObjectName([]string{"foo"})
	if ml != "foo" {
		t.Errorf("Expected foo")
	}
}

func TestParseVertex(t *testing.T) {
	v, err := parseVertex([]string{""})
	if err == nil {
		t.Errorf("Expected error")
	}

	v, err = parseVertex([]string{"1.000000 -1.000000"})
	if err == nil {
		t.Errorf("Expected error")
	}

	args := strings.Split("1.000000 -1.000000 -1.000000", " ")
	expected := Vec3{1, -1, -1}
	if v, err = parseVertex(args); err != nil {
		t.Errorf("Error: %v", err)
	} else if Vec3(v) != expected {
		t.Errorf("Expected vertex %v, found %v", expected, v)
	}
}

func TestParseFace(t *testing.T) {
	f, err := parseFace([]string{""})
	if err == nil {
		t.Errorf("Expected error")
	}

	f, err = parseFace([]string{"2 3"})
	if err == nil {
		t.Errorf("Expected error")
	}

	// Vertex only
	args := strings.Split("2 3 4", " ")
	expected := newFaceV([3]uint32{1, 2, 3})
	if f, err = parseFace(args); err != nil {
		t.Errorf("Error: %v", err)
	} else if *f != *expected {
		t.Errorf("Expected face %v, found %v", expected, f)
	}

	// Vertex/texture coord
	args = strings.Split("2/3 4/5 6/7", " ")
	expected = newFaceVT([3]uint32{1, 3, 5}, [3]uint32{2, 4, 6})
	if f, err = parseFace(args); err != nil {
		t.Errorf("Error: %v", err)
	} else if *f != *expected {
		t.Errorf("Expected face %v, found %v", expected, f)
	}

	// Vertex/texture coord/normal
	args = strings.Split("2/3/4 5/6/7 8/9/10", " ")
	expected = newFaceVTN([3]uint32{1, 4, 7}, [3]uint32{2, 5, 8}, [3]uint32{3, 6, 9})
	if f, err = parseFace(args); err != nil {
		t.Errorf("Error: %v", err)
	} else if *f != *expected {
		t.Errorf("Expected face %v, found %v", expected, f)
	}

	// Vertex//normal
	args = strings.Split("2//3 4//5 6//7", " ")
	expected = newFaceVN([3]uint32{1, 3, 5}, [3]uint32{2, 4, 6})
	if f, err = parseFace(args); err != nil {
		t.Errorf("Error: %v", err)
	} else if *f != *expected {
		t.Errorf("Expected face %v, found %v", expected, f)
	}
}

func TestParseObjCube(t *testing.T) {
	objStr := `
# Blender v2.71 (sub 0) OBJ File: ''
# www.blender.org
mtllib cube.mtl
o Cube
v 1.000000 -1.000000 -1.000000
v 1.000000 -1.000000 1.000000
v -1.000000 -1.000000 1.000000
v -1.000000 -1.000000 -1.000000
v 1.000000 1.000000 -0.999999
v 0.999999 1.000000 1.000001
v -1.000000 1.000000 1.000000
v -1.000000 1.000000 -1.000000
usemtl Material
s off
f 2 3 4
f 8 7 6
f 1 5 6
f 2 6 7
f 7 8 4
f 1 4 8
f 1 2 4
f 5 8 6
f 2 1 6
f 3 2 7
f 3 7 4
f 5 1 8`

	if _, err := parseObj(objStr, nil); err != nil {
		t.Errorf("Error parsing obj: %v", err.Error())
	}
}

func TestParseObjMaterial(t *testing.T) {
	matStr := `
# Blender MTL File: 'None'
# Material Count: 1

newmtl Material
Ns 96.078431
Ka 0.000000 0.000000 0.000000
Kd 0.640000 0.640000 0.640000
Ks 0.500000 0.500000 0.500000
Ni 1.000000
d 1.000000
illum 2`

	if _, err := parseMaterials(matStr); err != nil {
		t.Errorf("Error parsing obj material: %v", err.Error())
	}
}
