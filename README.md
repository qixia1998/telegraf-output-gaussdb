# telegraf-output-gaussdb
Write the timing data of the captured plug-in to GaussDB.

## Usage

#### git clone
``` 
git clone https://github.com/qixia1998/telegraf-output-gaussdb.git
```

#### build
```
go build -o gauss_output cmd/main.go
```

#### config
Edit `plugin.conf`.
```
[[outputs.gauss]]
  host = gauss_host
  port = gauss_port
  user = gauss_user
  password = gauss_password
  table = table_name
  debug = false
```

`telegraf.conf`<br>
inputs cpu example:
```
[[outputs.execd]]
  command = ["./gauss_output", "-config", "plugin.conf"]
  data_format = "influx"


[[outputs.influxdb_v2]]
  urls = ["your_url"]
  token = "your_token"
  organization = "your_org"
  bucket = "your_bucket"



# Read metrics about cpu usage
[[inputs.cpu]]
  ## Whether to report per-cpu stats or not
  percpu = true
  ## Whether to report total system cpu stats or not
  totalcpu = true
  ## If true, collect raw CPU time metrics
  collect_cpu_time = false
  ## If true, compute and report the sum of all non-idle CPU states
  ## NOTE: The resulting 'time_active' field INCLUDES 'iowait'!
  report_active = false
  ## If true and the info is available then add core_id and physical_id tags
  core_tags = false
  
```

#### run
```
telegraf --config telegraf.conf
```




