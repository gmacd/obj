package obj

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type SubMesh struct {
	MaterialName string
	Material     *Material
	Faces        []Face
}

type ObjMesh struct {
	name      string
	materials map[string]*Material

	Verts   []Vertex
	Normals []Vertex

	SubMeshes []*SubMesh
}

// We only support a subset of the parameters...
type Material struct {
	Name string
	//Specular       float32     // Ns
	//Ambient        core.Colour // Ka
	Diffuse Colour // Kd (this is what we really care about!)
	//SpecularColour core.Colour // Ks
	//RefactiveIndex float32     // Ni
}

type Vec3 [3]float32
type Colour [4]float32

// LoadObj loads the obj and material library using the provided func,
// and returns an ObjMesh pointer.
func LoadObj(objPath, mtlPath string, readFile func(path string) ([]byte, error)) *ObjMesh {
	// Load the obj and mtl library
	objBytes, err := readFile(objPath)
	if err != nil {
		panic(fmt.Sprintf("Couldn't load mesh file %v. Error:\n%v", objPath, err.Error()))
	}
	mtlBytes, err := readFile(mtlPath)
	if err != nil {
		panic(fmt.Sprintf("Couldn't load mesh file %v. Error:\n%v", mtlPath, err.Error()))
	}

	// We'll ignore the mtllib instruction in the obj, and give it this one
	var materials map[string]*Material
	if materials, err = parseMaterials(string(mtlBytes)); err != nil {
		panic(fmt.Sprintf("Couldn't parse file %v. Error:\n%v", mtlPath, err.Error()))
	}

	// Now parse the obj
	objData, err := parseObj(string(objBytes), materials)
	if err != nil {
		panic(fmt.Sprintf("Couldn't parse file %v. Error:\n%v", objPath, err.Error()))
	}

	return objData
}

type ObjectName string

type Vertex Vec3

type Face struct {
	v [3]uint32 // Vertex index
	t [3]uint32 // Texture coord index
	n [3]uint32 // Normal index
}

func newFaceV(v [3]uint32) *Face {
	return &Face{v, [3]uint32{0, 0, 0}, [3]uint32{0, 0, 0}}
}

func newFaceVT(v, t [3]uint32) *Face {
	return &Face{v, t, [3]uint32{0, 0, 0}}
}

func newFaceVN(v, n [3]uint32) *Face {
	return &Face{v, [3]uint32{0, 0, 0}, n}
}

func newFaceVTN(v, t, n [3]uint32) *Face {
	return &Face{v, t, n}
}

func parseObj(objStr string, materials map[string]*Material) (*ObjMesh, error) {
	mesh := &ObjMesh{}
	mesh.materials = materials

	var subMesh *SubMesh

	for _, line := range strings.Split(objStr, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		tokens := strings.Split(line, " ")
		if len(tokens) == 0 {
			continue
		}

		switch tokens[0] {
		case "#":
			// Ignore comments

		case "mtllib":
			// Do nothing - we're passing in a mtl file

		case "usemtl":
			materialName, err := parseMaterialName(tokens[1:])
			if err != nil {
				return nil, err
			}

			subMesh = &SubMesh{materialName, materials[materialName], nil}
			mesh.SubMeshes = append(mesh.SubMeshes, subMesh)

		case "o":
			objectName, err := parseObjectName(tokens[1:])
			if err != nil {
				return nil, err
			}
			mesh.name = objectName

		case "s":
			// Don't care

		case "v":
			v, err := parseVertex(tokens[1:])
			if err != nil {
				return nil, err
			}
			mesh.Verts = append(mesh.Verts, v)

		case "vn":
			v, err := parseVertex(tokens[1:])
			if err != nil {
				return nil, err
			}
			mesh.Normals = append(mesh.Normals, v)

		case "f":
			f, err := parseFace(tokens[1:])
			if err != nil {
				return nil, err
			}
			subMesh.Faces = append(subMesh.Faces, *f)

		default:
			return nil, fmt.Errorf("Unknown obj definition: (%v), %v", line, tokens)
		}
	}

	return mesh, nil
}

func parseMaterials(str string) (map[string]*Material, error) {
	matlib := make(map[string]*Material, 0)

	var mat *Material
	for _, line := range strings.Split(str, "\n") {
		line = strings.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		tokens := strings.Split(line, " ")
		if len(tokens) == 0 {
			continue
		}

		switch tokens[0] {
		case "#":
			// Ignore comments

		case "newmtl":
			mat = &Material{}
			mat.Name = tokens[1]
			matlib[mat.Name] = mat

		case "Kd":
			if values, err := parseFloat32Array(tokens[1:]); err != nil {
				return nil, err
			} else {
				mat.Diffuse[0] = values[0]
				mat.Diffuse[1] = values[1]
				mat.Diffuse[2] = values[2]
			}
		}
	}

	return matlib, nil
}

func parseMaterialName(args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("Expected only one argument")
	}
	return args[0], nil
}

func parseObjectName(args []string) (string, error) {
	if len(args) != 1 {
		return "", errors.New("Expected only one argument")
	}
	return args[0], nil
}

func parseVertex(args []string) (Vertex, error) {
	v := Vec3{0, 0, 0}
	if len(args) != 3 {
		return Vertex(v), fmt.Errorf("Expected only three arguments but found: %v, (%v)", args, len(args))
	}

	for i, _ := range args {
		f, err := strconv.ParseFloat(args[i], 32)
		if err != nil {
			return Vertex(v), fmt.Errorf("Couldn't parse element %v", i)
		}
		v[i] = float32(f)
	}

	return Vertex(v), nil
}

func parseFace(args []string) (*Face, error) {
	if len(args) != 3 {
		return nil, fmt.Errorf("Expected only three arguments but found: %v, (%v)", args, len(args))
	}

	v := [3]uint32{0, 0, 0}
	t := [3]uint32{0, 0, 0}
	n := [3]uint32{0, 0, 0}

	hasT, hasN := false, false

	for i, _ := range args {
		str := args[i]
		vStr, tStr, nStr := "", "", ""

		firstIdx := strings.Index(str, "/")
		secondIdx := -1
		if firstIdx != -1 {
			secondIdx = strings.Index(str[firstIdx+1:], "/")
			if secondIdx != -1 {
				secondIdx += firstIdx + 1
			}
		} else {
			secondIdx = len(str) - 1
		}

		if firstIdx == -1 {
			vStr = str
		} else {
			vStr = str[:firstIdx]
			if secondIdx == -1 {
				// Only texture coord
				tStr = str[firstIdx+1:]
			} else {
				// Normal and texture
				tStr = str[firstIdx+1 : secondIdx]
				nStr = str[secondIdx+1:]
			}
		}

		hasT = len(tStr) > 0
		hasN = len(nStr) > 0

		idx, err := strconv.ParseInt(vStr, 10, 32)
		if err != nil {
			return nil, fmt.Errorf("Couldn't parse vert %v, index %v", args, i)
		}
		// Face indices are 1-based - we'd rather them be 0-based
		v[i] = uint32(idx) - 1

		if hasT {
			idx, err = strconv.ParseInt(tStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("Couldn't parse tex %v, index %v", args, i)
			}
			// 1-base -> 0-base
			t[i] = uint32(idx) - 1
		}

		if hasN {
			idx, err = strconv.ParseInt(nStr, 10, 32)
			if err != nil {
				return nil, fmt.Errorf("Couldn't parse normal %v, index %v", args, i)
			}
			// 1-base -> 0-base
			n[i] = uint32(idx) - 1
		}
	}

	if !hasT && !hasN {
		return newFaceV(v), nil
	} else if hasT && !hasN {
		return newFaceVT(v, t), nil
	} else if !hasT && hasN {
		return newFaceVN(v, n), nil
	} else {
		return newFaceVTN(v, t, n), nil
	}
}

func parseFloat32Array(args []string) ([]float32, error) {
	v := make([]float32, 0)

	for i, _ := range args {
		f, err := strconv.ParseFloat(args[i], 32)
		if err != nil {
			return v, fmt.Errorf("Couldn't parse element %v", i)
		}
		v = append(v, float32(f))
	}

	return v, nil
}
