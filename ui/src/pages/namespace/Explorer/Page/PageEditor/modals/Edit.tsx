import {
  LayoutSchemaType,
  PageElementSchemaType,
  TableContentSchemaType,
  TextContentSchemaType,
} from "~/pages/namespace/Explorer/Page/PageEditor/schema";

import TableForm from "./forms/Table";
import TextForm from "./forms/Text";

const EditModal = ({
  layout,
  pageElementID,
  close,
  success,
  onChange,
}: {
  layout: LayoutSchemaType;
  pageElementID: number;
  close: () => void;
  success: (newLayout: LayoutSchemaType) => void;
  onChange: () => void;
}) => {
  const placeholder1: PageElementSchemaType = {
    name: "Text",
    hidden: false,
    content: { type: "Text", content: "This is a Text..." },
    preview: "This is a Text...",
  };

  const oldElement = layout?.[pageElementID] ?? placeholder1;

  const onEditText = (content: TextContentSchemaType) => {
    const newElement = {
      name: oldElement.name,
      hidden: oldElement.hidden,
      preview: content.content,
      content: { type: content.type, content: content.content },
    };

    const newLayout = [...layout];

    newLayout.splice(pageElementID, 1, newElement);

    success(newLayout);
    close();
  };

  const onEditTable = (content: TableContentSchemaType) => {
    const ObjectToString =
      content.content === undefined
        ? ""
        : content.content
            .map((element) => `${element.header}:${element.cell}, `)
            .join("");

    const newElement = {
      name: oldElement.name,
      hidden: oldElement.hidden,
      preview: ObjectToString,
      content,
    };

    const newLayout = [...layout];

    newLayout.splice(pageElementID, 1, newElement);

    success(newLayout);
    close();
  };

  const type = layout ? layout[pageElementID]?.name : "Text";

  return (
    <>
      {type === "Text" && (
        <TextForm
          layout={layout}
          onEdit={onEditText}
          pageElementID={pageElementID}
        />
      )}
      {type === "Table" && (
        <TableForm
          onChange={onChange}
          layout={layout}
          onEdit={onEditTable}
          pageElementID={pageElementID}
        />
      )}
    </>
  );
};

export default EditModal;
