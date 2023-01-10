import { Tooltip } from '@mui/material';
import MUIButton, { ButtonProps as MUIButtonProps } from '@mui/material/Button';
import { styled } from '@mui/material/styles';
import * as React from 'react';

export interface ButtonProps extends MUIButtonProps {
    /**
    * Tooltip to show on hover.
    */
    tooltip?: string;
    /**
    * Tooltip to show on hover when disabled is true. If unset, disabled tooltip will default to tooltip.
    */
    disabledTooltip?: string;
    asyncDisable?: boolean;
    /**
    *Auto expand height and width.
    */
    auto?: boolean;
    /**
    * Disables shadows on button.
    */
    disableShadows?: boolean;
    loading?: boolean; // ???
}

const DirektivButton = styled(MUIButton, {
    shouldForwardProp: (prop) => prop !== 'auto' && prop !== 'disabledTooltip' && prop !== 'disableShadows' && prop !== 'loading',
})<ButtonProps>(({ theme, color, size, auto, variant, disabledTooltip, disabled, disableShadows, loading }) => ({
    // Defaults
    textTransform: "none",
    fontSize: "0.8rem",
    padding: "0.4rem 0.5rem",
    minWidth: "auto",
    height: "auto",
    fontWeight: "bold",
    "&:visited": {
        color: color !== undefined && color !== "inherit" ? theme.palette[color].main : undefined
    },
    "&.MuiButton-sizeSmall": {
        height: "1.8rem",
        lineHeight: "1rem",
    },
    "&.MuiButton-sizeMedium": {
        height: "2.8rem",
        lineHeight: "2rem",
    },
    "&.MuiButton-sizeLarge": {
        height: "3.8rem",
        lineHeight: "3rem",
    },
    // Enable Shadows for non-text variants
    ...(variant !== "text" && {
        boxShadow: "var(--theme-shadow-box-shadow)",
        "&:hover": {
            backgroundColor: color !== undefined && color !== "inherit" ? theme.palette[color].light : undefined,
            transition: '0.2s'
        },
    }),
    // Custom Style for info color
    ...(color === "info" && variant !== "text" && {
        backgroundColor: "white",
        // outline: "var(--border)",
        borderColor: "var(--theme-subtle-border)",
        ":hover": {
            borderColor: "#d2d4d7"
        }
    }),
    // Custom Style for terminal color
    ...(color === "terminal" && variant !== "text" && {
        border: "none",
        ":disabled": {
            backgroundColor: "#2e3d48",
            color: "#65747f"
        },
    }),
    // Support Disabled Tooltips
    ...(disabledTooltip !== undefined && {
        "&:disabled": {
            pointerEvents: "auto",
        },
        ...(disabled && {
            "&:hover": {
                backgroundColor: undefined
            },
        }),
    }),
    // Auto expand height/width
    ...(auto && {
        width: "100%",
        minWidth: "0px",
        minHeight: "0px",
        height: "auto !important"
    }),
    ...(disableShadows && {
        boxShadow: undefined
    }),
}));

function Button({ tooltip, onClick, asyncDisable, disabledTooltip, disabled, ...props }: ButtonProps) {
    const [isOnClick, setIsOnClick] = React.useState(false)
    const tooltipText = React.useMemo(() => {
        const isDisabled = isOnClick || disabled
        if (isDisabled) {
            if (disabledTooltip !== undefined) {
                return disabledTooltip
            }

            return ""
        }

        return tooltip ? tooltip : ""
    }, [disabledTooltip, tooltip, isOnClick, disabled])

    return (
        <Tooltip title={tooltipText} placement="top" arrow>
                <DirektivButton variant="contained" color="primary" disableRipple size="small" {...props} disabled={isOnClick || disabled} disabledTooltip={disabledTooltip} onClick={async (e) => {
                    if (onClick === undefined) {
                        return
                    }

                    if (asyncDisable) {
                        setIsOnClick(true)
                    }

                    await onClick(e)

                    if (asyncDisable) {
                        setIsOnClick(false)
                    }

                }
                } />
        </Tooltip>
    )
}

export default Button