import { Eye, EyeOff } from "lucide-react";

import Button from "~/design/Button";
import { Card } from "~/design/Card";
import Input from "~/design/Input";
import { InputWithButton } from "~/design/InputWithButton";
import { useApiActions } from "~/util/store/apiKey";
import { useState } from "react";

const ApiKeyPanel = () => {
  const [apiKey, setApiKey] = useState("");
  const [showKey, setShowKey] = useState(false);
  const { setApiKey: storeApiKey } = useApiActions();

  return (
    <Card className="flex flex-col gap-5 p-5 md:flex-row ">
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

      <Button
        className="md:w-1/4"
        block
        disabled={!apiKey}
        onClick={() => {
          storeApiKey(apiKey);
        }}
      >
        set API key
      </Button>

      <Button
        className="md:w-1/4"
        variant="destructive"
        onClick={() => {
          storeApiKey(null);
        }}
      >
        unset API key
      </Button>
    </Card>
  );
};

export default ApiKeyPanel;
