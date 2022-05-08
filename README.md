# Dead Rabbit

---

Dead rabbit is a tool for investigating RabbitMQ DLQ

Before use, you should prepare configuration file, following below syntax. Configuration file should be called 'configuration.yaml' and present in the "current" directory, from which you're running DeadRabbit 

```yaml
debug: false
rabbitmq:
  host: "<string>" # RabbitMQ Host
  port: "<number>" # RabbitMQ Port
  user: "<string>" # RabbitMQ Username
  password: "<string>" # RabbitMQ Password
  vhost: "<string>" # RabbitMQ Virtual host
  queue: "<string>" # RabbitMQ Queue to resend messages
  dlq: "<string>" # RabbitMQ Dead letter queue, to read messages from
databases:
  - name: Finance # DB Name, shown in list, following by query name; You could specify more than 1 db
    host: "<string>" # DB Host
    port: "<number>" # DB Port
    user: "<string>" # DB Username
    password: "<string>" # DB Password
    schema: "<string>" # DB schema to use
    queries:
      - name: "Select LineItems by id" # Name, displayed in a list
        # Below you can observe a query "format" string. Query can be parametrized. 
        # Query parameters are specified in a following format ":<Parameter name>"
        # Each <Parameter name> should be described in 'params' dictionary with same 'params.name'
        format: >
          SELECT *
          FROM LINE_ITEMS
          WHERE id = :id
        params: # Known bug: parameters list shouldn't be empty
          - name: id
            # Each parameter will be substituted in query using fmt.Sprintf(format, value)
            # value got from user's input in corresponding input field in dialog
            # value is always a string value
            format: "%s"
```
