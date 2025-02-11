import { LayoutSchemaType } from "~/pages/namespace/Explorer/Page/PageEditor/schema";
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
  const oldElement = layout ? layout[pageElementID] : { content: "nothing" };

  const onEdit = (content: unknown) => {
    let newElement;
    if (type === "Table") {
      // const ObjectToString =
      //   content.content === undefined
      //     ? ""
      //     : content.content.map(
      //         (element) => `${element.header}:${element.cell}, `
      //       );

      const ObjectToString = "something";

      newElement = {
        name: oldElement?.name,
        hidden: oldElement?.hidden,
        preview: ObjectToString,
        content,
      };
    } else {
      newElement = {
        name: oldElement?.name,
        hidden: oldElement?.hidden,
        preview: content,
        content,
      };
    }

    const newLayout = [...layout];

    newLayout.splice(pageElementID, 1, newElement);

    success(newLayout);
    close();
  };

  const type = layout ? layout[pageElementID]?.name : "nothing";

  return (
    <>
      {type === "Text" && (
        <TextForm
          layout={layout}
          onEdit={onEdit}
          pageElementID={pageElementID}
        />
      )}
      {type === "Table" && (
        <TableForm
          onChange={onChange}
          layout={layout}
          onEdit={onEdit}
          pageElementID={pageElementID}
        />
      )}
    </>
  );
};

export default EditModal;
