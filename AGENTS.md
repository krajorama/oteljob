# General

Create reports in Markdown format.

Always verify the report by running the given use case in docker compose. Flag if the use case is not possible to run.

In result tables, be explicit about the result, always include the actual result for OTel model, Prometheus exposition, Prometheus time series. One exception is if something is fully byte-to-byte equal to something else, then you can say identical to xy. Show differences with bold letters, do not say things like "identical except xy" or "same as before plus".

It is ok to add short explanation after the results. Longer explanations should go into itemized footnotes.

# Measurement points

1. OTel model in the application.
2. Prometheus series exposed for scrape in the application via OTel SDK exporter.
3. OTel model created from the exposition by the OTel Collector Prometheus receiver.
4. Time series received in Prometheus via the OTel collector Prometheus remote write exporter.
