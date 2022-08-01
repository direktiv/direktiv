import { createTheme } from '@mui/material/styles';

const Colors = {
    primary: "#3E94C5",
    secondary: "#95defb",
    light: "#566875"
};

const theme = createTheme({
    typography: {
        fontFamily: [
            "Inter"
        ],
        fontSize: "14px",
        fontWeight: "bold",
    },
    palette: {
        primary: {
            main: Colors.primary
        },
        secondary: {
            main: Colors.secondary
        },
        text: {
            primary: Colors.light
        }
    },
    components: {
        // Name of the component
        MuiPaginationItem: {
            defaultProps: {
                disableRipple: true,
            }
        }
    },
});

export default theme;