import { createTheme } from '@mui/material/styles';
import * as React from 'react';
import {
    Link as RouterLink,
    LinkProps as RouterLinkProps
} from 'react-router-dom';

const Colors = {
    primary: "#3E94C5",
    secondary: "#95defb",
    light: "#566875"
};

// React Router link behavior for MUIButton
const LinkBehavior = React.forwardRef<
  HTMLAnchorElement,
  Omit<RouterLinkProps, 'to'> & { href: RouterLinkProps['to'] }
>((props, ref) => {
  const { href, ...other } = props;
  return <RouterLink data-testid="custom-link" ref={ref} to={href} {...other} />;
});

const theme = createTheme({
    typography: {
        fontFamily: [
            "Inter"
        ].join(","),
        body1: {
            fontWeight: "bold",
            fontSize: "14px"
        }
    },
    palette: {
        primary: {
            main: Colors.primary,
            light: "#44a3d9"
        },
        secondary: {
            main: Colors.secondary
        },
        text: {
            primary: Colors.light
        },
        info: {
            main: "#566875",
            light: "#e6e6e6"
        },
        error: {
            main: "#ffc0c4"
        },
        terminal: {
            main: "#355166",
            light: "#3a5970",
        }
    },
    components: {
        // Name of the component
        MuiPaginationItem: {
            defaultProps: {
                // disableRipple: true,
            }
        },
        MuiTooltip: {
            styleOverrides: {
                tooltipArrow: {
                    backgroundColor: "#1a3041"
                },
                arrow: {
                    color: "#1a3041"
                }
            }
        },
        MuiButtonBase: {
            defaultProps: {
              LinkComponent: LinkBehavior,
            },
        },
    },
});

declare module '@mui/material/styles' {
    interface Palette {
        terminal: Palette['primary'];
    }

    // allow configuration using `createTheme`
    interface PaletteOptions {
        terminal?: PaletteOptions['primary'];
    }
}

// Update the Button's color prop options
declare module '@mui/material/Button' {
    interface ButtonPropsColorOverrides {
        terminal: true;
    }
}

export default theme;