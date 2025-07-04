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
  return <DraggableElement blockPath={[1]} element={block} id="1" />;
};
