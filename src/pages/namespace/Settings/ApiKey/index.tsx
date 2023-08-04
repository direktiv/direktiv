import { Eye, EyeOff } from "lucide-react";
import { useApiActions, useApiKey } from "~/util/store/apiKey";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { useState } from "react";

const ApiKeyPanel = () => {
  const [showKey, setShowKey] = useState(false);
  const { setApiKey: storeApiKey } = useApiActions();
  const apiKeySore = useApiKey();
  const [apiKey, setApiKey] = useState(apiKeySore ?? "");

  return (
    <Card>
      <form
        className="flex flex-col gap-5 p-5"
        action=""
        onSubmit={(e) => {
          e.preventDefault();
          storeApiKey(apiKey);
        }}
      >
        <InputWithButton className="w-full">
          <Input
            value={apiKey}
            onChange={(e) => {
              setApiKey(e.target.value);
            }}
            type={showKey ? "text" : "password"}
            placeholder="enter API key"
          />

          <Button
            icon
            variant="ghost"
            onClick={() => {
              setShowKey((prev) => !prev);
            }}
          >
            {showKey ? <EyeOff /> : <Eye />}
          </Button>
        </InputWithButton>

        <Button block type="submit" disabled={!apiKey}>
          set API key
        </Button>
        <Button
          variant="destructive"
          onClick={() => {
            storeApiKey(null);
          }}
        >
          unset API key
        </Button>
      </form>
    </Card>
  );
};

export default ApiKeyPanel;
