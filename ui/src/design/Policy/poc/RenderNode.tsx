import { AndGroup, Connector, OrGroup } from "../Group";
import { type PolicyLayoutNode, shouldRenderConnector } from "./utils";
import { Condition } from "~/design/Policy/Condition";
import { Placeholder } from "~/design/Policy/Placeholder";

const RenderNode = ({ node }: { node: PolicyLayoutNode }) => {
  // Leaf nodes are the terminal expressions in the policy
  // tree, they render directly as a condition component
  if (node.type === "leaf") {
    return (
      <Condition className="font-mono" title={node.title}>
        <span className="block w-full truncate">{node.preview}</span>
      </Condition>
    );
  }

  // OR nodes render as vertical branches. The trailing
  // placeholder creates an extra empty branch
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

  // AND nodes render children in a linear sequence. Each item expands to either
  // [node] or [node, connector], and flatMap combines those into one React
  // children array, e.g. [A, Connector, B, Connector, C]. A trailing connector
  // and placeholder may be appended afterward.
  return (
    <AndGroup>
      {node.items.flatMap((item, index) => {
        const nextItem = node.items[index + 1];
        const renderedItems = [<RenderNode key={index} node={item} />];

        // Connect neighboring AND siblings unless either side is an
        // OR group. OR groups render their own branching structure
        if (nextItem !== undefined && shouldRenderConnector(item, nextItem)) {
          renderedItems.push(<Connector key={index} />);
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
