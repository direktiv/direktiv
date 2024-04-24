import { isEnterprise } from "~/config/env/utils";
import { useApiKey } from "~/util/store/apiKey";
import { useAuthTest } from "~/api/authenticate/query/getAuthInfos";

/**
 * This hook will provide information about api key handling with the
 * following properties:
 *
 * isApiKeyRequired: indicates if the api needs an api key to work.
 * In enterprise mode this is always true.
 *
 * isCurrentKeyValid: indicates if this stored key from the user can be
 * successfully used to authenticate against the api. The user might not
 * have a stored key and this test will then be run with an undefined key
 *
 * isFetched: indicates if the api key handling is finished. As long as
 * this is false, isApiKeyRequired and isCurrentKeyValid can be undefined
 *
 * showLoginModal: a flag that indicates whether the UI should handle the
 * displaying a login modal.
 *
 * showUsermenu: indicates whether the usermenu should be shown. In the
 * enterprise version this is always true (and is independent from any
 * api key handling), in the open source version this is only true if
 * the api key is required
 */
const useApiKeyHandling = () => {
  const storedKey = useApiKey();
  const keyIsPresent = !!storedKey;

  const { data: storedKeyTestResult, isFetched: testWithStoredKeyFinished } =
    useAuthTest({
      apikey: storedKey ?? undefined,
      enabled: keyIsPresent,
    });

  const isApiKeyRequired = window?._direktiv?.requiresAuth ?? false;

  const isCurrentKeyValid = keyIsPresent
    ? storedKeyTestResult
    : !isApiKeyRequired;

  const isFetched = keyIsPresent ? testWithStoredKeyFinished : true;

  return {
    isApiKeyRequired,
    isCurrentKeyValid,
    isFetched,
    showLoginModal: !isEnterprise(),
    showUsermenu: isEnterprise() ? true : isApiKeyRequired,
  };
};

export default useApiKeyHandling;
