import RefreshButton from ".";
import { useState } from "react";

export default {
  title: "Components/RefreshButton",
  parameters: { layout: "fullscreen" },
};

export const Default = () => (
  <div className="p-5">
    <RefreshButton icon size="sm" variant="ghost" />
  </div>
);

export const DisabledButton = () => (
  <div className="p-5">
    <RefreshButton icon size="sm" variant="ghost" disabled />
  </div>
);

export const ButtonWithTextVariation = () => (
  <div className="flex gap-5 p-5">
    <RefreshButton size="lg" variant="primary">
      Large Primary Button
    </RefreshButton>
    <RefreshButton variant="outline">Medium Outline</RefreshButton>
    <RefreshButton variant="link">Link</RefreshButton>
  </div>
);

export const Circle = () => (
  <div className="flex gap-5 p-5">
    <RefreshButton circle icon />
    <RefreshButton circle icon disabled />
  </div>
);

export const OnClickDemo = () => {
  const [count, setCount] = useState(0);
  return (
    <div className="flex flex-col gap-5 p-5">
      <div>
        This demo shows that the onCLick event is forwared to the RefreshButton
        correctly. The Counter is <span className="font-bold">{count}</span>
      </div>
      <div>
        <RefreshButton
          variant="outline"
          onClick={() => {
            setCount((old) => old + 1);
          }}
        >
          Refresh Counter
        </RefreshButton>
      </div>
    </div>
  );
};
