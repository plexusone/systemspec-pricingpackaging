# Writing a spec

A Spec is a JSON document. It is read with unknown fields rejected, so a typo is
an error, not silently ignored. Its JSON Schema is `schema/pricing.schema.json`,
for editors.

## A small offer

```json
{
  "formatVersion": "1",
  "app": "taskboard",
  "currency": "usd",
  "defaultPlan": "free",
  "applications": [ { "id": "boards", "name": "Boards" } ],
  "tiers": [
    { "id": "free", "name": "Free", "rank": 1 },
    { "id": "plus", "name": "Plus", "rank": 2 }
  ],
  "features": [
    { "id": "support", "name": "Support", "levels": ["community", "email"] }
  ],
  "limits": [
    { "id": "boards", "name": "Boards", "unit": "board", "period": "none" }
  ],
  "plans": [
    {
      "id": "free", "name": "Free", "tier": "free", "selfServe": true,
      "features": [ { "feature": "support", "level": "community" } ],
      "limits": [ { "limit": "boards", "max": 3 } ]
    },
    {
      "id": "plus", "name": "Plus", "tier": "plus", "selfServe": true, "trialDays": 14,
      "prices": [
        { "interval": "month", "amountCents": 1200 },
        { "interval": "year", "amountCents": 12000 }
      ],
      "features": [ { "feature": "support", "level": "email" } ],
      "limits": [ { "limit": "boards", "max": 25 } ]
    }
  ],
  "entitlements": [
    { "tier": "free", "application": "boards", "seats": 3 },
    { "tier": "plus", "application": "boards", "seats": 10 }
  ]
}
```

Fuller offers are in `examples/`: a subscription tool with levelled features and
fair-use allowances, and a marketplace with seller-set prices and per-plan fees.

## Rules worth knowing

- **Amounts are integers** in the currency's minor units: `1200` is $12.00.
- **Zero means zero.** A limit a plan does not set, or sets to `0`, allows
  nothing. Write `"unlimited": true` for no limit.
- **Ladders are only for levels that include the one below.** A set of
  independent options (video from one provider, from another) is separate
  switches, not levels.
- **Fair use.** With a `max`, exceeding it warns and flags the account for review
  instead of refusing. With `unlimited`, it records that the allowance is
  advertised as unlimited and subject to your fair-use policy.
- **The default plan is free and self-serve,** and its tier needs a grant, so an
  account that is not billed can still open something.
- **Every other self-serve plan has a price.** An arranged plan (`selfServe`
  false) may have none.
- **Seats live only in grants.** To limit members, change the tier's grant.
- **A marketplace fee is in basis points** (1000 is 10%), and a plan's
  `platformFeeBps` overrides it for sellers on that plan.

## Checking it

```go
f, _ := os.Open("pricing.json")
spec, err := pricingspec.Load(f)
if err != nil {
	var ve *pricingspec.ValidationError
	if errors.As(err, &ve) {
		for _, is := range ve.Issues {
			fmt.Println(is) // E006 error: plans[1].features[0].level: ...
		}
	}
	return err
}
for _, w := range spec.Lint() {
	fmt.Println(w) // W002 warning: ...
}
```

Run it in CI so an invalid offer cannot ship. Lint warnings are for a person to
look at; they never reject a Spec.
