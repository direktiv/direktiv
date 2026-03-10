import {
  ElbowContext,
  generateElbowPathParams,
  generateElbowPaths,
} from "./paths";

import { getSum } from "../utils";

type ElbowComponentProps = {
  targets: number[];
  reverse?: boolean;
  rowHeight?: number;
  width?: number;
  r?: number;
};

/**
 * Generates elbow connectors. See Policy stories in storybook for examples.
 *  
 * @param {ElbowComponentProps} props
 * @param {number[]} props.targets - the size (number of rows) of each child group.
 * @param {boolean} [props.reverse=false] - mirrors the elbows to get closing bracket (right to left)
 * @param {number} [props.rowHeight=96] - optional, only for experimentation in storybook - height in px of one row
 * @param {number} [props.width=64] - optional, only for experimentation in storybook - width of the elbows element
 * @param {number} [props.radius=12] - optional, only for experimentation in storybook - radius of the round corners

 * @returns {ReactNode} an SVG with the elbow connector paths
 */
const Elbows = ({
  targets,
  reverse = false,
  rowHeight = 96,
  width = 64,
  r = 12,
}: ElbowComponentProps): JSX.Element => {
  const heightUnits = getSum(targets);
  const totalHeight = rowHeight * heightUnits;
  const startX = reverse ? width : 0;
  const startY = totalHeight / 2;
  const endX = reverse ? 0 : width;
  const xm = width / 2;
  const xl = reverse ? width / 2 + r : (width - r * 2) / 2;
  const xr = reverse ? (width - r * 2) / 2 : width / 2 + r;

  // these coordinates are the same for all elbow paths, e.g. all points on the x axis.
  const elbowContext: ElbowContext = {
    startX, // the x axis position of the common start point for all elbows.
    startY, // the y axis position of the common start point for all elbows.
    endX, // the endpoint for all elbows on the x axis.
    xm, // the middle line
    xl, // the position left of the middle line, where the rounded corner arc starts.
    xr, // the position right of the middle line, where the rounded corner arc ends.
    r, // the radius for the rounded corners
    reverse, // whether to draw "closing" elbows from right to left
  };

  const elbowPathParams = generateElbowPathParams(
    targets,
    startY,
    rowHeight,
    r
  );

  const elbowPaths = generateElbowPaths(elbowPathParams, elbowContext);

  return (
    <svg
      viewBox={`0 0 ${width} ${totalHeight}`}
      className="fill-none stroke-gray-400 stroke-2"
      style={{ width }}
    >
      {...elbowPaths}
    </svg>
  );
};

export { Elbows };
