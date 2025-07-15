## Description

This PR outlines how form primitives (such as text inputs, select boxes, etc.) are defined in the Direktiv Pages spec and how they interact with their form block. It also extends the form block as the current specification only has limited data type supported in the payload of the request.

Forms represent one of the most challenging and critical features for the Direktiv Pages. As a feature it requires maximum flexibility, forms must support a wide range of use cases and data flexibility.

We distinguish between a form block (the block that actually sends a request to the backend) and form primitives (form element blocks like an input field. These are nested however deep somewhere in the form block.

This is how a page with a form could be structured like:

![page builder form](https://github.com/user-attachments/assets/fe2a7e33-60bc-49e3-86f5-59c11630802c)

## Form primitives

Each form primitive should be pretty straightforward. They are simple blocks that add a variable to the page's state and can be used in the form block. Each of these elements return a specific data type: text areas store strings, checkboxes store booleans, and so on.

### All form elements should have these fields in common

- `id`: for referencing it in the form block
- `label`
- `description`: a string that gives a more details description about the form element. It can be an empty string.
- `defaultValue`: an optional default value of that form element. The type of that value must match the output type
- `required`: a boolean that determines if this field is required to be set by the user

### I suggest we use the following primitives:

- input `string` or `number`
  - additional attributes:
    - type: `text` | `data` | `password` | `email` | `url` | `number`
- textarea `string`
- checkbox `boolean`
- select `string`
  - additional attributes:
    - values: `string[]`. This attribute will list all the possible select values that are available.
- file `string` (for future iteration)
  - Does not need to be implemented in the first iteration, but it will definitely be a requirement at some point. File input usually has some implications about the content type of the request and therefore I wanted to make sure that we can solve this data type conceptually already. Right now, we imply that forms are sent via `Content-Type: application/json`. However, a binary is not compatible with JSON and therefore is mostly sent via `Content-Type: multipart/form-data`. This would mean that using a file input in a form would implicitly change the entire payload from `Content-Type: application/json` to `Content-Type: multipart/form-data`. I would suggest starting with a simple implementation first and always casting the file to a base64 encoded string and sending it via `Content-Type: application/json` as well. I think most admins will own the backend API and should be able to handle file uploads that way. However, I can imagine that there are cases where the admin does not own the upload endpoint like if it is a presigned URL for an S3 bucket, but if this becomes a requirement we can still implement a different behavior.
- checkbox list (or a multi select) `string[]` (for future iteration)

### primitives that we don't need

- I decided against radio buttons, as a select box can represent the same state in a much more compact way

## A new data type we need

For mutations and queries, we use the `KeyValue` schema that looks like this:

```
{
  "key": "some-key",
  "value": "some value that supports template strings {{query.user.id}}"
}
```

This data type is perfect for modelling request params and request headers as the key and values can only be strings. However the payload of a mutation must be 100% JSON compatible which requires a new data type. To solve this I would like to introduce [a new Schema](https://github.com/direktiv/direktiv/pull/1850/files#diff-03d7ea036d1551a6f15a40de724d88e61e5099a9a4d5c69465a2f50cfe4a74c1).

Here is an example of a complex request that is now possible with the new schema. This request could be made from the form in the picture above:

```JavaScript
{
  "mutation": {
    "id": "create-ticket",
    "url": "/api/teams/{{query.user.teamId}}/projects/{{loop.project.id}}/tickets",
    "method": "POST",
    "queryParams": [
      {
        "key": "assigned",
        "value": "{{query.user.id}}"
      }
    ],
    "requestHeaders": [
      {
        "key": "Authorization",
        "value": "Bearer {{query.user.token}}"
      }
    ],
    // request body must support more that just key value string pairs
    "requestBody": [
      {
        "key": "title",
        "value": {
          "type": "string",
          // a string using a variable placeholder from a string input
          "value": "Draft: {{form.ticketForm.title}}"
        }
      },
      {
        "key": "description",
        "value": {
          "type": "string",
          // a static string
          "value": "Steps to reproduce: \n\n Acceptance criteria: \n"
        }
      },
      {
        "key": "priority",
        "value": {
          "type": "variable",
          // uses a variable and preserves type. In this
          // it would be sourced from a number input
          "value": "form.ticketForm.priority"
        }
      },
      {
        "key": "hidden",
        "value": {
          "type": "variable",
          // boolean value from a checkbox
          "value": "form.ticketForm.hidden"
        }
      },
      {
        "key": "isDraft",
        "value": {
          "type": "boolean",
          // a static boolean
          "value": true
        }
      },
      {
        "key": "categories",
        "value": {
          "type": "variable",
          // this is an example of using a variables
          // that does not come from a form at all
          "value": "loop.project.categories"
        }
      },
      {
        "key": "relatedTickets",
        "value": {
          "type": "array",
          // a static array of strings
          "value": ["ticket-1", "ticket-2", "ticket-3"]
        }
      },
      {
        "key": "customFields",
        "value": {
          "type": "object",
          // a static object
          "value": [
            {
              "key": "severity",
              "value": "high"
            },
            {
              "key": "environment",
              "value": "staging"
            }
          ]
        }
      }
    ]
  }
}
```

This will compile to

```bash
curl -X POST "https://example.com/api/teams/team-123/projects/proj-789/tickets?assigned=user-456" \
  -H "Authorization: Bearer abc123" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Draft: Login button not responsive",
    "description": "Steps to reproduce: \n\n Acceptance criteria: \n",
    "priority": 2,
    "hidden": false,
    "isDraft": true,
    "categories": ["bug", "frontend"],
    "relatedTickets": ["ticket-1", "ticket-2", "ticket-3"],
    "customFields": {
      "severity": "high",
      "environment": "staging"
    }
  }'
```

This new data type would also solve a similar problem that we have in the form elements.

Every form element has a `defaultValue` attribute. If these could just be strings, we could use a template string here and would cover all possible default values. However, we could not set a default value for a checkbox (`boolean`), a number input (`number`) or a checkbox list (`string[]`). Setting a default value could now look like this

```javascript
// setting a string
"defaultValue": {
  "type": "string",
  "value": "Draft: {{loop.project.title}}"
}
```

or

```javascript
// setting a number (assuming the variable stores a number)
"defaultValue": {
  "type": "variable",
  "value": "query.project.defaultPriority"
}
```

or

```javascript
// setting a static number
"defaultValue": {
  "type": "number",
  "value": 2
}
```

We will have a similar problem and solution for defining all the values that a select can have. We want the ability to set it from a static list of strings

```javascript
"value": {
  "type": "array",
  value: ["low", "medium", "high", "urgent"],
}
```

but also source them from a variable.

```javascript
"value": {
  "type": "variable",
  "value": "query.project.availablePriorities"
}
```

## Future Improvements

- **Validation**: With the ability to mark form elements as required or optional, we already have a very pragmatic validation that is suitable for an MVP. However, in the future, we need more granular validation. This should be defined in the form component very close to the mutation, maybe even in the `RequestBodySchema`.

## Specific Changes in this PR:

- modeled the new schema for the payload
- disabled the form payload for now as it now requires a new implementation (not disabling it would cause TypeScript errors)
  - user can not set the payload anymore in the UI
  - even if a payload is set, the request will ignore it for now

## Checklist

- [x] Documentation updated if required
- [x] Test coverage is appropriate

## Checklist Internal

- [x] Linear issue linked (e.g. [DIR-XXXX] pull request title)
- [x] Has the PR been labeled
