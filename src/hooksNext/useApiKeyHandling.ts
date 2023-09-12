import env from "~/config/env";
import { useApiKey } from "~/util/store/apiKey";
import { useAuthTest } from "~/api/authenticate/query/getAuthInfos";

/**
 * Send test request to check if api needs auth. In enterprise
 * mode, this test will be skipped and will always return false
 * because the api authentication will not be managed by the react
 * app
 */
const useIsAuthRequired = () => {
  const { VITE_IS_ENTERPRISE: isEnterprise } = env;
  const { data: testSucceeded, isFetched: isFnished } = useAuthTest({
    enabled: !isEnterprise,
  });
  return isEnterprise
    ? { authRequired: false, isFinished: true }
    : { authRequired: !testSucceeded, isFinished: isFnished };
};

/**
 * this hook will test, whether the the api needs an auth key by sending a
 * request. It also tests, whether the stored key is valid (if there is one)
 */
const useApiKeyHandling = () => {
  const storedKey = useApiKey();
  const keyIsPresent = !!storedKey;

  const { data: storedKeyTestResult, isFetched: testWithStoredKeyFinished } =
    useAuthTest({
      apikey: storedKey ?? undefined,
      enabled: keyIsPresent,
    });

  const { authRequired, isFinished: authCheckFinished } = useIsAuthRequired();

  return {
    isKeyRequired: authRequired,
    isCurrentKeyValid: keyIsPresent ? storedKeyTestResult : !authRequired,
    isFetched: keyIsPresent
      ? authCheckFinished && testWithStoredKeyFinished
      : authCheckFinished,
  };
};

export default useApiKeyHandling;
