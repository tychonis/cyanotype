BUILD := build
CACHE := cache

.PHONY: all clean

CY := cyanotype
ROOT := {ROOT}
SERVER := {SERVER}

# Lazy evaluation for BOM and GLBS.
BOM = $(shell $(CY) bom $(ROOT) --format list)
GLBS = $(addprefix $(BUILD)/,$(addsuffix .glb,$(BOM)))

all: $(GLBS)

$(BUILD)/%.glb: $(CACHE)/%.cyscene
	@mkdir -p $(BUILD)
	brepkit glb $< $@
	$(CY) scene upload $(SERVER) $* $@

.SECONDARY:
$(CACHE)/%.cyscene:
	@mkdir -p $(CACHE)
	$(CY) scene generate $* -o $@

clean:
	rm -rf $(BUILD) $(CACHE)
