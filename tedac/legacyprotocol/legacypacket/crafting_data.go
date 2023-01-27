package legacypacket

import (
	"fmt"
	"github.com/sandertv/gophertunnel/minecraft/protocol"
	"github.com/sandertv/gophertunnel/minecraft/protocol/packet"
	"github.com/tedacmc/tedac/tedac/legacyprotocol"
)

// CraftingData is sent by the server to let the client know all crafting data that the server maintains. This
// includes shapeless crafting, crafting table recipes, furnace recipes etc. Each crafting station's recipes
// are included in it.
type CraftingData struct {
	// Recipes is a list of all recipes available on the server. It includes among others shapeless, shaped
	// and furnace recipes. The client will only be able to craft these recipes.
	Recipes []legacyprotocol.Recipe
	// PotionRecipes is a list of all potion mixing recipes which may be used in the brewing stand.
	PotionRecipes []protocol.PotionRecipe
	// PotionContainerChangeRecipes is a list of all recipes to convert a potion from one type to another,
	// such as from a drinkable potion to a splash potion, or from a splash potion to a lingering potion.
	PotionContainerChangeRecipes []protocol.PotionContainerChangeRecipe
	// ClearRecipes indicates if all recipes currently active on the client should be cleaned. Doing this
	// means that the client will have no recipes active by itself: Any CraftingData packets previously sent
	// will also be discarded, and only the recipes in this CraftingData packet will be used.
	ClearRecipes bool
}

// ID ...
func (*CraftingData) ID() uint32 {
	return packet.IDCraftingData
}

// Marshal ...
func (pk *CraftingData) Marshal(w *protocol.Writer) {
	l := uint32(len(pk.Recipes))
	w.Varuint32(&l)
	for _, recipe := range pk.Recipes {
		var c int32
		switch recipe.(type) {
		case *legacyprotocol.ShapelessRecipe:
			c = protocol.RecipeShapeless
		case *legacyprotocol.ShapedRecipe:
			c = protocol.RecipeShaped
		case *legacyprotocol.FurnaceRecipe:
			c = protocol.RecipeFurnace
		case *legacyprotocol.FurnaceDataRecipe:
			c = protocol.RecipeFurnaceData
		case *legacyprotocol.MultiRecipe:
			c = protocol.RecipeMulti
		case *legacyprotocol.ShulkerBoxRecipe:
			c = protocol.RecipeShulkerBox
		case *legacyprotocol.ShapelessChemistryRecipe:
			c = protocol.RecipeShapelessChemistry
		case *legacyprotocol.ShapedChemistryRecipe:
			c = protocol.RecipeShapedChemistry
		default:
			w.UnknownEnumOption(fmt.Sprintf("%T", recipe), "crafting recipe type")
		}
		w.Varint32(&c)
		recipe.Marshal(w)
	}
	protocol.Slice(w, &pk.PotionRecipes)
	protocol.Slice(w, &pk.PotionContainerChangeRecipes)
	w.Bool(&pk.ClearRecipes)
}

// Unmarshal ...
func (pk *CraftingData) Unmarshal(r *protocol.Reader) {
	var length uint32
	r.Varuint32(&length)
	pk.Recipes = make([]legacyprotocol.Recipe, length)
	for i := uint32(0); i < length; i++ {
		var recipeType int32
		r.Varint32(&recipeType)

		var recipe legacyprotocol.Recipe
		switch recipeType {
		case protocol.RecipeShapeless:
			recipe = &legacyprotocol.ShapelessRecipe{}
		case protocol.RecipeShaped:
			recipe = &legacyprotocol.ShapedRecipe{}
		case protocol.RecipeFurnace:
			recipe = &legacyprotocol.FurnaceRecipe{}
		case protocol.RecipeFurnaceData:
			recipe = &legacyprotocol.FurnaceDataRecipe{}
		case protocol.RecipeMulti:
			recipe = &legacyprotocol.MultiRecipe{}
		case protocol.RecipeShulkerBox:
			recipe = &legacyprotocol.ShulkerBoxRecipe{}
		case protocol.RecipeShapelessChemistry:
			recipe = &legacyprotocol.ShapelessChemistryRecipe{}
		case protocol.RecipeShapedChemistry:
			recipe = &legacyprotocol.ShapedChemistryRecipe{}
		default:
			r.UnknownEnumOption(recipeType, "crafting data recipe type")
		}
		//goland:noinspection GoNilness
		recipe.Unmarshal(r)
		pk.Recipes[i] = recipe
	}
	protocol.Slice(r, &pk.PotionRecipes)
	protocol.Slice(r, &pk.PotionContainerChangeRecipes)
	r.Bool(&pk.ClearRecipes)
}
