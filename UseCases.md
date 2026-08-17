1. A standalone application is using the OTel SDK with Prometheus exporter enabled. The exporter is using default translation (underscores and suffixes).

In the application the OTel model is:
Resource{service.name="my_service", service.instance.id="my_id"}
Scope{name="foo", version=""}
Metric{name="received_samples", type=Sum, monotonic=true, temporality=cumulative}
DataPoint{attributes={"a.a":"b"}, value=1}

2. A standalone application is using the OTel SDK with Prometheus exporter enabled. The exporter is using no translation (utf-8).

In the application the OTel model is:
Resource{service.name="my_service", service.instance.id="my_id"}
Scope{name="foo", version=""}
Metric{name="received_samples", type=Sum, monotonic=true, temporality=cumulative}
DataPoint{attributes={"a.a":"b"}, value=1}

3. An observer application is observing multiple (two) applications.

App 1 OTel model:
Resource{service.name="my_service1", service.instance.id="my_id1"}
Scope{name="foo", version=""}
Metric{name="received_samples", type=Sum, monotonic=true, temporality=cumulative}
DataPoint{attributes={"a.a":"b"}, value=1}

App 2 OTel model:
Resource{service.name="my_service2", service.instance.id="my_id2"}
Scope{name="foo", version=""}
Metric{name="received_samples", type=Sum, monotonic=true, temporality=cumulative}
DataPoint{attributes={"a.a":"b"}, value=2}

In this case the exporter must add job, instance labels to series exposed about the application:

All series (foo and target_info) generated for App 1 must have job="my_service1" and instance="my_id1" labels.
All series (foo and target_info) generated for App 2 must have job="my_service2" and instance="my_id2" labels.
