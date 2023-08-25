import { useApiKey } from "~/util/store/apiKey";
import { useAuthTest } from "~/api/authenticate/query/getAuthInfos";

/**
 * this hook will test, whether the the api needs an auth key by sending a
 * request. It also tests, whether the stored key is valid (if there is one)
 */
const useApiKeyHandling = () => {
  const apiKeyFromLocalstorage = useApiKey();
  const apiKeyisPresent = !!apiKeyFromLocalstorage;

  const { data: testWithStoredKey, isFetched: testWithStoredKeyFinished } =
    useAuthTest({
      apikey: apiKeyFromLocalstorage ?? undefined,
      enabled: apiKeyisPresent,
    });

  const { data: testWithoutKey, isFetched: testWithoutKeyFinished } =
    useAuthTest();

  return {
    isKeyRequired: !testWithoutKey,
    isCurrentKeyValid: apiKeyisPresent ? testWithStoredKey : testWithoutKey,
    isSuccess: apiKeyisPresent
      ? testWithoutKeyFinished && testWithStoredKeyFinished
      : testWithoutKeyFinished,
  };
};

export default useApiKeyHandling;
