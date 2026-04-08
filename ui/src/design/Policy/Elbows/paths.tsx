type ElbowPathParam = {
  s: number;
  targetY: number;
  endY: number;
  ya: number;
  yb: number;
};

export type ElbowContext = {
  endX: number;
  startX: number;
  startY: number;
  xl: number;
  xm: number;
  xr: number;
  reverse: boolean;
  r: number;
};

/**
 * Receives an array "targets" with the height (in rows) of each elbow target,
 * and generates the params required to define an SVG path.
 *
 * @param targets - e.g., [1, 1, 3]
 * @param startY - y coordinate for the common start point on the left border of the component.
 * @param rowHeight - height of one row; corresponds to height of one Condition element + margin.
 * @param r - radius for the arcs of the rounded corners.
 * @returns {ElbowPathParam}
 */
export const generateElbowPathParams = (
  targets: number[],
  startY: number,
  rowHeight: number,
  r: number
) =>
  targets.reduce<ElbowPathParam[]>((acc, ySize) => {
    // get previous endpoint y position from accumulator or start at 0.
    const previousEndY = acc[acc.length - 1]?.endY ?? 0;

    // the end of the current row on the y axis: previous end point + current row height
    const endY = previousEndY + ySize * rowHeight;

    // the target point y position, in the middle of the current row
    const targetY = endY - (ySize * rowHeight) / 2;

    // s is positive if the elbow goes downward, negative if upward.
    const s = targetY > startY ? 1 : -1;

    const newElbow = {
      s,
      endY,
      targetY,
      ya: startY + r * s, // point on the y axis where left rounded corner arc ends
      yb: targetY - r * s, // point on y axis where right rounded corner arc starts
    };

    return [...acc, newElbow];
  }, []);

/**
 * Receives all relevant coordinates for drawing the SVG paths for the elbows.
 *
 * @param params - array with relevant params for the individual elbows (for points on y axis)
 * @param context - object describing the common params for all elbows (mostly x axis values)
 * @returns
 */
export const generateElbowPaths = (
  params: ElbowPathParam[],
  context: ElbowContext
) => {
  const { endX, startX, startY, xl, xm, xr, reverse, r } = context;

  return params.map((item, index) => {
    if (startY === item.targetY) {
      // If start and end point are on same y, just draw straight line.
      return (
        <line key={index} x1={startX} y1={startY} x2={endX} y2={item.targetY} />
      );
    }

    // Otherwise, draw elbow with rounded corners.

    // Below, each segment of the svg path is defined in a separate line
    // for better readability. These are later concatenated into one <path> element.

    // The comments describe how to draw the elbow, step by step.
    // Each step starts at the x,y coordinate reached at the end of the previous step.

    // Go to the start point (x,y).
    const startPoint = `M${startX},${startY}`;

    // draw line the right until xl is reached
    const horizontalL = `H ${xl}`;

    // draw an arc that ends in xm and item.ya
    const arcL = `A ${r} ${r} 0 0 ${
      (!reverse && item.s === 1) || (reverse && item.s === -1) ? 1 : 0
    } ${xm} ${item.ya}`;

    // draw a vertical line up (or down) to item.yb
    const vertical = `V ${item.yb}`;

    // draw an arc that ends in xr and item.targetY
    const arcR = `A ${r} ${r} 0 0 ${
      (!reverse && item.s === 1) || (reverse && item.s === -1) ? 0 : 1
    } ${xr} ${item.targetY}`;

    // draw a horizontal line to the right until reaching endX
    const horizontalR = `H ${endX}`;

    // merge strings into one path
    const d = [startPoint, horizontalL, arcL, vertical, arcR, horizontalR].join(
      "\n"
    );

    return <path key={index} d={d} />;
  });
};
