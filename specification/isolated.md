# Isolated

## Input

Your function's input data can be found at `/direktiv-data/input.json`. 

## Logging 

Write logs to `/direktiv-data/out.log`. You should probably create this file on startup, even if you don't log anything.

Logs should be written to this file. They will be read in line-by-line by Direktiv.

## Responding 

If the function executes correctly you must write the results to `/direktiv-data/output.json`. Standard rules apply. This will usually be JSON, but can be something else, which will be base64 encoded by direktiv.

## Errors

If you want to report that something has gone wrong, write a JSON object to `/direktiv-data/error.json` with the `code` and `msg` keys. Same rules as other functions: any provided `code` makes the error catchable, otherwise it's uncatchable. 

Example:

```
{
	"code": "badInput",
	"msg": "Should be a bool, not an int."
}
```

## Finishing

It is critical that you create a file called `/direktiv-data/done` when you're finished. You don't need to write anything into it, but from the moment it's created you're in a race to the finish. The sidecar will kick in and report the results and terminate the container. You should either report an Error or Respond with results before creating this file.

After creating this file, you should exit with exit status zero. 
