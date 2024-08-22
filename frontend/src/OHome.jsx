import React from "react";
import { Box, Button, Text, Flex, Image } from "@chakra-ui/react";
import { BrowserRouter as Router, Route, Routes, Link } from "react-router-dom";
import backgroundImage from "./assets/images/bgnf.jpg";
import logo from "./assets/images/neuralforge.png";


// Define your components for different routes here
const Home = () => <Text>Home Page</Text>;
const AboutMe = () => <Text>About Me Page</Text>;

class OHome extends React.Component {
  render() {
    return (
      <Box>
        <Router>
          <Flex
            direction="column"
            minHeight="100vh"
            position="relative"
            boxShadow="0 0 60px 60px rgba(0, 0, 0, 0.9)" // Strong shadow to eliminate white edges
            _before={{
              content: '""',
              position: "absolute",
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              backgroundImage: `url(${backgroundImage})`, // Reference the imported image
              backgroundSize: "cover",
              backgroundRepeat: "no-repeat",
              backgroundPosition: "center",
              zIndex: "-1",
            }}
            bg="rgba(255, 255, 255, 0.2)" // Semi-transparent white background for the glass effect
            backdropFilter="blur(6px)" // Apply blur to the background for the glass effect
          >
            <Box flex="1" position="relative" zIndex="1">
              <Box position="relative" zIndex="1" paddingBottom="100px">
                <Routes>
                  <Route path="/" element={<Home />} />
                  <Route path="/project" element={<AboutMe />} />
                  {/* Add more routes as needed */}
                </Routes>
              </Box>
            </Box>
            <Box
              as="footer"
              width="100%"
              backgroundColor="rgba(255, 255, 255, 0.2)" // Bring back the glassy effect on the footer
              color="white"
              padding="10px"
              textAlign="center"
            >
              <Flex align="center">
                <Link to="/">
                  <Image
                    src={logo}
                    alt="Home"
                    boxSize="40px"
                    cursor="pointer"
                    mr="4"
                    _hover={{
                      transform: "scale(1.1)", // Slightly enlarge the logo
                      transition: "transform 0.2s ease-in-out", // Smooth transition
                    }}
                  />
                </Link>
                <Link to="/project">
                  <Button colorScheme="teal" variant="outline" mr="4">
                    Project
                  </Button>
                </Link>
              </Flex>
            </Box>
          </Flex>
        </Router>
      </Box>
    );
  }
}

export default OHome;
