import { useApiKey } from "~/util/store/apiKey";
import { useAuthTest } from "~/api/authenticate/query/getAuthInfos";

const isEnterprise = process.env.VITE?.VITE_IS_ENTERPRISE === "true";

/**
 * Send test request to check if api needs an api key. In enterprise
 * mode, this test will be skipped and will always return false
 * because the api authentication will not be managed by the react
 * app
 */
const useIsApiKeyRequired = () => {
  const { data: testSucceeded, isFetched: isFinished } = useAuthTest({
    enabled: !isEnterprise,
  });
  return isEnterprise
    ? { isApiKeyRequired: false, isFinished: true }
    : {
        isApiKeyRequired:
          testSucceeded === undefined ? undefined : !testSucceeded,
        isFinished,
      };
};

/**
 * This hook will provide information about api key handling with the
 * following properties:
 *
 * isApiKeyRequired: indicates if the api needs an api key to work.
 * In enterprise mode this is always false, because the login in handled
 * by a different layer in front of the api/UI
 *
 * isCurrentKeyValid: indicates if this stored key from the user can be
 * successfully used to authenticate against the api. The user might not
 * have a stored key and this test will then be run with an undefined key
 *
 * isFetched: indicates if the api key handling is finished. As long as
 * this is false, isApiKeyRequired and isCurrentKeyValid can be undefined
 *
 * showUsermenu: indicates whether the usermenu should be shown. In the
 * enterprise version this is always true (and is independent from any
 * api key handling), in the open source version this is only true if
 * the api is required
 */

const useApiKeyHandling = () => {
  const storedKey = useApiKey();
  const keyIsPresent = !!storedKey;

  const { data: storedKeyTestResult, isFetched: testWithStoredKeyFinished } =
    useAuthTest({
      apikey: storedKey ?? undefined,
      enabled: keyIsPresent,
    });

  const { isApiKeyRequired, isFinished: authCheckFinished } =
    useIsApiKeyRequired();

  const isCurrentKeyValid = keyIsPresent
    ? storedKeyTestResult
    : !isApiKeyRequired;

  const isFetched = keyIsPresent
    ? authCheckFinished && testWithStoredKeyFinished
    : authCheckFinished;

  return {
    isApiKeyRequired,
    isCurrentKeyValid,
    isFetched,
    showUsermenu: isEnterprise ? true : isApiKeyRequired,
  };
};

export default useApiKeyHandling;
