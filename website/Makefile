HUGO         ?= hugo
SCREENSHOTS  := ../screenshots
LOGO_SVG     := ../frontend/public/logo-full.svg
IMG_DST      := static/images

.PHONY: setup serve build clean

setup:
	@mkdir -p $(IMG_DST)/screenshots
	@cp $(SCREENSHOTS)/*.png $(IMG_DST)/screenshots/ 2>/dev/null && echo "Screenshots copied." || echo "No screenshots found — skipping."
	@cp $(LOGO_SVG) $(IMG_DST)/logo-full.svg 2>/dev/null && echo "Logo copied." || echo "Logo not found — skipping."

serve: setup
	$(HUGO) server --disableFastRender --buildDrafts

build: setup
	$(HUGO) --minify

clean:
	rm -rf public/ resources/ $(IMG_DST)/screenshots/ $(IMG_DST)/logo-full.svg
