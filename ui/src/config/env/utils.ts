const defaultIsEnterpriseValue = false;

const isEnterpriseEnvValue = process.env.VITE?.VITE_IS_ENTERPRISE;

/**
 *
 * isEnterprise is determined by the following order of precedence:
 * - the value of the VITE_IS_ENTERPRISE environment variable,
 *   if it is set to either "true", "TRUE", "false" or "FALSE"
 * - the value of the window._direktiv.isEnterprise property, if
 *   it is a boolean
 * - the content of the defaultIsEnterprise variable*
 */
export const isEnterprise = () => {
  if (isEnterpriseEnvValue !== undefined) return isEnterpriseEnvValue;

  let isEnterPriseWindowValue = defaultIsEnterpriseValue;
  if (
    typeof window !== "undefined" &&
    typeof window._direktiv?.isEnterprise === "boolean"
  ) {
    isEnterPriseWindowValue = window._direktiv.isEnterprise;
  }
  return isEnterPriseWindowValue;
};
