REPORTS_MD  := $(wildcard verification-*.md)
REPORTS_HTML := $(REPORTS_MD:.md=.html)

.PHONY: html clean-html
html: $(REPORTS_HTML)

%.html: %.md report-style.html
	pandoc -s --embed-resources --standalone \
		--metadata title="$$(sed -n '/^# /{s/^# //;s/[`*_]//g;p;q}' $<)" \
		-H report-style.html \
		-o $@ $<

clean-html:
	rm -f $(REPORTS_HTML)
