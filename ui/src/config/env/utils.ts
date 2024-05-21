export const isEnterprise = () => {
  let isEnterprise = false;
  if (
    typeof window !== "undefined" &&
    typeof window._direktiv?.isEnterprise === "boolean"
  ) {
    isEnterprise = window._direktiv.isEnterprise;
  }
  return isEnterprise;
};
