import { AllBlocksType } from "~/pages/namespace/Explorer/Page/poc/schema/blocks";
import { DraggableElement } from "./index";

export default {
  title: "Components/DragAndDropEditor/DraggableElement",
};

export const Default = () => {
  const block: AllBlocksType = {
    type: "button",
    label: "Button",
  };
  return <DraggableElement element={block} name="1" />;
};
