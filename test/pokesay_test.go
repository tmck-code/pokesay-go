package test

import (
	"testing"

	"github.com/tmck-code/pokesay/src/pokedex"
	"github.com/tmck-code/pokesay/src/pokesay"
)

func TestListCategories(test *testing.T) {
	t := pokedex.NewTrie()
	t.Insert([]string{"small", "g1", "r"}, pokedex.NewPokemonEntry(0, "pikachu"))
	t.Insert([]string{"small", "g1", "o"}, pokedex.NewPokemonEntry(1, "bulbasaur"))
	t.Insert([]string{"medium", "g1", "o"}, pokedex.NewPokemonEntry(2, "bulbasaur"))
	t.Insert([]string{"big", "g1", "o"}, pokedex.NewPokemonEntry(3, "bulbasaur"))
	t.Insert([]string{"big", "g1"}, pokedex.NewPokemonEntry(4, "charmander"))

	result := pokesay.ListCategories(*t)
	assert([]string{"big", "g1", "medium", "o", "r", "small"}, result, result, test)
}
