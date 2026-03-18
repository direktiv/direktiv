import { AndGroup, Connector, OrGroup } from "../Group";
import { type NodeVM, shouldRenderConnector } from "./utils";
import { Condition } from "~/design/Policy/Condition";
import { Placeholder } from "~/design/Policy/Placeholder";

const RenderNode = ({ node }: { node: NodeVM }) => {
  if (node.type === "leaf") {
    return (
      <Condition className="font-mono" title={node.title}>
        <span className="block w-full truncate">{node.preview}</span>
      </Condition>
    );
  }

  if (node.type === "or") {
    return (
      <OrGroup childSizes={[...node.childSizes, 1]}>
        {node.branches.map((branch, index) => (
          <RenderNode key={index} node={branch} />
        ))}
        <AndGroup>
          <Placeholder />
        </AndGroup>
      </OrGroup>
    );
  }

  return (
    <AndGroup>
      {node.items.flatMap((item, index) => {
        const nextItem = node.items[index + 1];
        const renderedItems = [
          <RenderNode key={`item-${index}`} node={item} />,
        ];

        if (nextItem !== undefined && shouldRenderConnector(item, nextItem)) {
          renderedItems.push(<Connector key={`connector-${index}`} />);
        }

        return renderedItems;
      })}
      {node.items.length > 0 && node.items.at(-1)?.type !== "or" && (
        <Connector />
      )}
      <Placeholder />
    </AndGroup>
  );
};

export default RenderNode;
