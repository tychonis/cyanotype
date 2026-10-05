BUILD := build
CACHE := cache
FAILED := $(BUILD)/failed.txt

.PHONY: all clean prepare

CY := cyanotype
ROOT := {ROOT}
SERVER := {SERVER}

# Lazy evaluation for BOM and GLBS.
BOM = $(shell $(CY) bom $(ROOT) --format list)
GLBS = $(addprefix $(BUILD)/,$(addsuffix .glb,$(BOM)))

all: prepare $(GLBS)
	@if [ -f $(FAILED) ]; then \
		echo; \
		echo "===== FAILED TARGETS ====="; \
		cat $(FAILED); \
		echo "=========================="; \
		exit 1; \
	fi

prepare:
	@mkdir -p $(BUILD)
	@rm -f $(FAILED)

$(GLBS): | prepare

$(BUILD)/%.glb: $(CACHE)/%.cyscene
	@mkdir -p $(BUILD)
	@echo "brepkit glb $< $@"
	@if brepkit glb $< $@; then \
		echo "$(CY) scene cache $(SERVER) $* $@"; \
		$(CY) scene cache $(SERVER) $* $@ || \
			echo "$* (cache)" >> $(FAILED); \
	else \
		echo "$* (build)" >> $(FAILED); \
		rm -f $@; \
	fi

.SECONDARY:
$(CACHE)/%.cyscene:
	@mkdir -p $(CACHE)
	$(CY) scene generate $* -o $@

clean:
	rm -rf $(BUILD) $(CACHE)
