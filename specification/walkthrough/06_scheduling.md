# Scheduling 

Sometimes you want a workflow to run periodically. Direktiv supports scheduling based on "cron", and in this article you'll see how that's done.

## Demo 

```yaml
id: scraper
start:
  type: scheduled
  cron: "0 */2 * * *"
functions:
- id: httprequest
  type: reusable
  image: vorteil/request
states:
- id: getter 
  type: action
  action:
    function: httprequest
    input: '{
      "method": "GET",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
    }'
  transform: '.return'
  transition: storer
- id: storer
  type: action 
  action:
    function: httprequest
    input: '{
      "method": "POST",
      "url": "https://jsonplaceholder.typicode.com/todos/1",
      "body": .
    }'
```

## Start Types

Workflow definitions can have one of many different start types. Up until now you've left the `start` section out entirely, which causes it to `default`, which is appropriate for a direct-invoke/subflow workflow. Now we can have a look at `scheduled` workflows.

```yaml
start:
  type: scheduled
  cron: "0 */2 * * *"
```

There's not much to see here. Add the `start` section, set `type` to `scheduled`, and define a valid cron string and away you go!

Direktiv prevents scheduled workflows from being directly invoked or used as a subflow, which is why this demo doesn't specify any input data. Just configure the workflow and check the logs over time to see the scheduled workflow in action.

## Active/Inactive Workflows

Every workflow definition can be considered "active" or "inactive". Being "active" doesn't mean that there's an instance running right now, it means that Direktiv will allow instances to be created from it. This setting is part of the API, not a part of the workflow definition.

With scheduled workflows we can finally see why this setting could be useful: you can toggle the schedule on and off without modifying the workflow definition itself.

## Cron

Cron is a time-based job scheduler in Unix-like operating systems. Direktiv doesn't run cron, but it does borrow their syntax and expressions for scheduling. 

In the example above our cron expression is "`0 */2 * * *`". This tells Direktiv to run the workflow once every two hours. There are many great resources online to help you create your own custom cron expressions.
