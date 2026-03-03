# Cedar JSON Policy Schema TODO

Track implementation progress for `https://docs.cedarpolicy.com/policies/json-format.html` in Zod.

## Policy object primitives

- [x] `effect` primitive
- [x] `principal` primitive
- [x] `action` primitive
- [x] `resource` primitive
- [x] `conditions` primitive
- [x] `annotations` primitive

## Shared policy primitives

- [x] `entity` primitive (`{ type, id }`)
- [x] `slot` primitive(s) (`?principal`, `?resource`)
- [x] operator primitives (`All`, `==`, `in`, `is`)

## JsonExpr primitives

- [x] `Value`
- [x] `Var`
- [x] `Slot`
- [x] `Unknown`
- [x] unary operators (`!`, `neg`, `isEmpty`)
- [x] binary operators (`==`, `!=`, `in`, `<`, `<=`, `>`, `>=`, `&&`, `||`, `+`, `-`, `*`, `contains`, `containsAll`, `containsAny`, `hasTag`, `getTag`)
- [x] attribute operators (`.`, `has`)
- [x] `is`
- [x] `like`
- [x] `if-then-else`
- [x] `Set`
- [x] `Record`
- [x] extension function/method calls (any other key)

## Policy set primitives (optional scope, same Cedar page)

- [x] `staticPolicies`
- [x] `templates`
- [x] `templateLinks`

## Tests

- [x] basic schema tests for `effect` and `principal`
- [x] add tests for `action`
- [x] add tests for `resource`
- [x] add tests for `conditions`
- [x] add tests for `annotations`
- [x] add tests for policy set schema
- [x] add focused tests for each JsonExpr primitive/operator
- [x] add invalid-shape tests for each discriminated/union variant
