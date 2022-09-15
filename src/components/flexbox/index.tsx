import React, { useMemo } from 'react';
import './style.css';

export type FlexBoxCenterAxis  = "y" | "x" | "xy" | boolean;
export type FlexBoxGapSize  = "md" | "sm" | boolean;

export interface FlexBoxProps extends React.HTMLAttributes<HTMLDivElement> {
    /**
    * Hides component if true.
    */
    hide?: boolean
    /**
    * Set flex direction to column.
    */
    col?: boolean
    /**
    * Set flex direction to row.
    */
    row?: boolean
    /**
    * If true or "xy" aligns children to the center on both the X and Y axis.
    * Alternatively children can be aligned along a single axis, by using "x" or "y".
    */
    center?: FlexBoxCenterAxis
    /**
    * If true and "md" space children with a gap. Alternatively the gap can be smaller, by using "sm".
    */
    gap?: FlexBoxGapSize
    /**
    * Sets component height to 100%
    */
    tall?: boolean
    /**
    * Enables wrapping on children components.
    */
    wrap?: boolean
}

/**
* A flex display based div component. It's a basic layout element designed to speed to the use of flexbox css properties.
*/
function FlexBox({
    hide,
    col,
    row,
    center,
    gap,
    tall,
    wrap ,
    className ,
    ...props
}: FlexBoxProps){
    const classes = useMemo(() => {
        const prefix = "flex-box"
        let clsName = className ? className : ""

        if (hide) {
            clsName += ` hide`
        }

        if (col) {
            clsName += ` col`
        }

        if (row) {
            clsName += ` row`
        }

        if (tall) {
            clsName += ` tall`
        }

        if (wrap) {
            clsName += ` wrap`
        }

        if (gap) {
            switch (gap) {
                case true: {
                    clsName += ` gap-md`
                    break;
                }
                case "md": {
                    clsName += ` gap-md`
                    break;
                }
                case "sm": {
                    clsName += ` gap-sm`
                    break;
                }
            }
        }

        if (center) {
            switch (center) {
                case true: {
                    clsName += ` center`
                    break;
                }
                case "xy": {
                    clsName += ` center`
                    break;
                }
                case "x": {
                    clsName += ` center-x`
                    break;
                }
                case "y": {
                    clsName += ` center-y`
                    break;
                }
            }
        }

        return `${prefix} ${clsName}`
    }, [className, hide, col, row, gap, center, tall, wrap])

    return (
        <div {...props} className={classes} />
    );
}

export default FlexBox;