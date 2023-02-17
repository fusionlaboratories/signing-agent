package crypto

import (
	"math/rand"
	"testing"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func StringWithCharset(length int, charset string) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func String(length int) string {
	return StringWithCharset(length, charset)
}

func GenerateTreeInputs(num_leaves int) (tree_leaves [][]byte, err error) {
	// Merkle tree builder needs  an array of []byte values
	for i := 0; i < num_leaves; i++ {

		// Will use IDDocs/ SignedAssets in E2E integrationtests
		// Create a dummy IDDoc (SignedAsset)
		//input, err := NewIDDoc("merkle_leaves")
		//if err != nil {
		//	return nil, err
		//}
		//input.AddTag("qredo_test_tag1", []byte("abc2"))
		//input.Sign(input)

		// SerializedSignedAsset is a method (attahced to SignedAssets types)
		// That marshals a SignedAsset into a byte slice
		//serialized_doc, err := input.SerializeSignedAsset()

		//Generate some random bytes
		//token := make([]byte, 8)
		//rand.Read(token)
		// Set SignedAsset seed to random bytes
		//input.Seed = token
		// Add SignedAsset to set of Merkle Tree leaves (inputs)

		// Create random string
		random_string := String(20)
		// Use bytes as leaves in Merkle tree
		random_bytes := []byte(random_string)

		tree_leaves = append(tree_leaves, random_bytes)
	}
	return tree_leaves, nil
}

func TestMerkleTreeBuilder(t *testing.T) {
	// Generate random tree inputs
	leaf_num := 4
	leaves, err := GenerateTreeInputs(leaf_num)
	if err != nil {
		t.Log(err)
		t.Fail()
	}

	// Build Merkle Tree
	tree, err := BuildMerkleTreeStore(leaves)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	// Test lenth of tree array
	nextPoT := nextPowerOfTwo(leaf_num)
	Size := nextPoT*2 - 1
	if len(tree) != Size {
		t.Error("Incorrect Tree array size.")
		t.Fail()
	}
}

// Test Merkle proof generator given some random input bytes
func TestMerkleProofGeneratorSpeed(t *testing.T) {
	// This test mesures the tree generation, proof extraction and verification speed.
	// For each loop the tree size (number of leaves) is set.
	// A tree is generated with random leaves, a proof for every element is extracted and verfied
	// The time taken to complete each cycle is logged.
	max := 512
	for i := 1; i < max; i++ {
		leaf_num := i
		leaves, err := GenerateTreeInputs(leaf_num)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		// Create Merkle Tree
		tree, err := BuildMerkleTreeStore(leaves)
		if err != nil {
			t.Log(err)
			t.Fail()
		}
		// Test 1: Generate and verify proof for every element of the tree
		for j := 0; j < len(leaves); j++ {
			proof, err := GenerateProofFromTree(tree[j], j, tree)
			if err != nil {
				t.Log(err)
				t.Fail()
			}
			root := tree[len(tree)-1]
			if proof[len(proof)-1] != root {
				t.Logf("Proof operator did not generate Merkle root.")
				t.Fail()
			}
			// We need to make a copy of the proof for verify
			proof_copy, err := CopyProof(proof)
			if err != nil {
				// Cannot copy proof
				t.Log(err)
				t.Fail()
			}
			// verify proof
			err = Verify(*root, proof_copy)
			if err != nil {
				// Proof has failed
				t.Log(err)
				t.Fail()
			}
		}
	}
}
func TestMerkleProofGeneratorError(t *testing.T) {
	leaf_num := 8
	leaves, err := GenerateTreeInputs(leaf_num)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	// Create Merkle Tree
	tree, err := BuildMerkleTreeStore(leaves)
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	// Test 2 Incorrect Position for element
	_, err = GenerateProofFromTree(tree[1], 0, tree)
	if err == nil {
		t.Error("Should return error. Wrong position was supplied.")
		t.Fail()
	}

	// Test 3 Element not in tree
	RandBytes := []byte("This is random.")
	_, err = GenerateProofFromTree(&RandBytes, 2, tree)
	if err == nil {
		t.Error("Should return error. argument 0 (RandBytes) is not an element of Merkle Tree (tree).")
		t.Fail()
	}

}
