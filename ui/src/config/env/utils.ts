const defaultIsEnterprise = false;

const isEnterpriseEnvVar = process.env.VITE?.VITE_IS_ENTERPRISE;

const isEnterpriseWindowVar =
  typeof window === "undefined"
    ? defaultIsEnterprise
    : !!window._direktiv?.isEnterprise;

export const isEnterprise =
  isEnterpriseEnvVar !== undefined ? isEnterpriseEnvVar : isEnterpriseWindowVar;
