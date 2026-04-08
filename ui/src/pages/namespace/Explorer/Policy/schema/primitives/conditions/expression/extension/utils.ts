// Good-enough decimal validation for this schema layer: require a Cedar-like
// decimal string shape, but leave overflow and exact bound checks out.
export const isValidDecimalLiteral = (value: string) =>
  /^-?\d+\.\d{1,4}$/.test(value);

// Good-enough duration validation for this schema layer: keep Cedar's unit
// order and basic string shape, but leave overflow checks out.
export const isValidDurationLiteral = (value: string) => {
  if (value === "" || value === "-") {
    return false;
  }

  return /^-?(?:\d+d)?(?:\d+h)?(?:\d+m)?(?:\d+s)?(?:\d+ms)?$/.test(value);
};
