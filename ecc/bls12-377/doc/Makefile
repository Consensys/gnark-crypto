# Generic Makefile for LaTeX compilation of scientific articles

SHELL := /bin/bash # And not merely "sh", because we use features like toto{ext1,ext2}

# Useful aliases:
FIGDIR := fig
PLTDIR := plot
IMGDIR := images

TEX_FILES := $(wildcard *.tex)
FIG_FILES := $(wildcard $(FIGDIR)/*.fig)
PLT_FILES := $(wildcard $(PLTDIR)/*.plot)
IMG_FILES := $(wildcard $(IMGDIR)/*.png)

# Images to be constructed for the document
FIG_PDF := $(patsubst %.fig,%.pdf,$(FIG_FILES))
PLT_PDF := $(patsubst %.plot,%.pdf,$(PLT_FILES))
IMG_PDF := $(patsubst %.png,%.pdf,$(IMG_FILES))
PDF     := $(FIG_PDF) $(PLT_PDF) $(IMG_PDF)

# The real targets
PDF_FILES := $(patsubst %.tex,%.pdf,$(TEX_FILES))

# Temporary files - Only needed by the clean target
DOC_AUX := $(patsubst %.tex,%.aux,$(TEX_FILES))
DOC_DVI := $(patsubst %.tex,%.dvi,$(TEX_FILES))
DOC_LOG := $(patsubst %.tex,%.log,$(TEX_FILES))
DOC_TOC := $(patsubst %.tex,%.toc,$(TEX_FILES))
DOC_NAV := $(patsubst %.tex,%.nav,$(TEX_FILES))
DOC_OUT := $(patsubst %.tex,%.out,$(TEX_FILES))
DOC_SNM := $(patsubst %.tex,%.snm,$(TEX_FILES))
DOC_VRB := $(patsubst %.tex,%.vrb,$(TEX_FILES))
DOC_BLG := $(patsubst %.tex,%.blg,$(TEX_FILES))
DOC_BBL := $(patsubst %.tex,%.bbl,$(TEX_FILES))
DOC_FLS := $(patsubst %.tex,%.fls,$(TEX_FILES))
DOC_LOF := $(patsubst %.tex,%.lof,$(TEX_FILES))
DOC_LOT := $(patsubst %.tex,%.lot,$(TEX_FILES))
DOC_LOA := $(patsubst %.tex,%.loa,$(TEX_FILES))
DOC_FDB := $(patsubst %.tex,%.fdb_latexmk,$(TEX_FILES))
DOC_SGZ := $(patsubst %.tex,%.synctex.gz,$(TEX_FILES))
DOC_BAK := $(patsubst %.tex,%.tex.bak,$(TEX_FILES))

# Defining ``all'' as the first target!
.PHONY: all clean
all: $(PDF) $(PDF_FILES)

clean:
	rm -f \
	$(DOC_AUX) \
	$(DOC_LOG) \
	$(DOC_TOC) \
	$(DOC_NAV) \
	$(DOC_OUT) \
	$(DOC_SNM) \
	$(DOC_VRB) \
	$(DOC_BLG) \
	$(DOC_BBL) \
	$(DOC_DVI) \
	$(DOC_FLS) \
	$(DOC_FDB) \
	$(DOC_SGZ) \
	$(DOC_LOT) \
	$(DOC_LOF) \
	$(DOC_LOA) \
	$(DOC_BAK) \
	missfont.log \
	*~

$(FIGDIR)/%.pdf: $(FIGDIR)/%.fig
	cd $(FIGDIR); ./fig2pdf.sh $(notdir $*); convert -density 300% $(notdir $*).{pdf,png}

$(PLTDIR)/%.pdf: $(PLTDIR)/%.plot
	cd $(PLTDIR); gnuplot $(notdir $<)

$(IMGDIR)/%.pdf: $(IMGDIR)/%.png
	convert $< $@

PDFLATEX = pdflatex -shell-escape #-sPDFPassword=password
%.pdf: %.tex
	aspell check -t -d english $<
	$(PDFLATEX) $<
	@# If the command hangs, note the error line, type 'Q', and correct the source
	if cat $*.tex | sed 's/%.*$$//' | grep -q '\\bibliography{'; then \
		bibtex $*; \
		iconv -f ISO-8859-15 -t utf-8 $*.bbl > $*.bbl.utf8; \
		mv $*.bbl{.utf8,}; \
	fi
	$(PDFLATEX) $< >/dev/null </dev/null
	$(PDFLATEX) $< >/dev/null </dev/null
