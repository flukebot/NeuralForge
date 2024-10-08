import React from "react";
import {
  Box,
  Button,
  Input,
  Text,
  VStack,
  Heading,
  List,
  ListItem,
  Switch,
  FormControl,
  FormLabel,
} from "@chakra-ui/react";
import axios from "axios";
import { Link } from "react-router-dom";

import { CreateProject, ListProjects } from "../../wailsjs/go/main/App";

class OSound extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      projectName: "",
      selectedProject: null,
      projects: [], // List of existing projects
      useWails: true, // Toggle for using Wails or Axios
      isSupervised: true, // Toggle for supervised or unsupervised
    };
  }

  componentDidMount() {
    this.loadProjects();
  }

  loadProjects = async () => {
    try {
      // Attempt to dynamically import and use Wails ListProjects
      if (this.state.useWails) {
        const wailsModule = await import("../../wailsjs/go/main/App");
        const ListProjects = wailsModule.ListProjects;

        ListProjects()
          .then((projects) => {
            console.log("Projects loaded using Wails");
            this.setState({
              projects: projects.length
                ? projects.filter((project) => project.startsWith("ns_"))
                : [],
            });
          })
          .catch((error) => {
            console.error(
              "Failed to load projects from Wails, trying Axios:",
              error
            );
            this.loadProjectsWithAxios(); // Fallback to Axios if Wails fails
          });
      } else {
        this.loadProjectsWithAxios(); // Directly use Axios if Wails is not available
      }
    } catch (error) {
      console.log("Server mode or Wails module failed to load:", error);
      this.loadProjectsWithAxios(); // Fallback to Axios if dynamic import fails
    }
  };

  loadProjectsWithAxios = async () => {
    const url = `http://${window.location.hostname}:8080/api/list-projects`;
    try {
      const response = await axios.get(url);

      if (response.status === 200) {
        console.log("Projects fetched from server:", response.data);
        this.setState({
          projects: response.data.length
            ? response.data.filter((project) => project.startsWith("ns_"))
            : [],
        });
      } else if (response.status === 204) {
        console.log("No projects found.");
        this.setState({ projects: [] });
      } else {
        console.error("Unexpected response status:", response.status);
      }
    } catch (error) {
      console.error("Failed to load projects using Axios:", error.message);
    }
  };

  handleProjectNameChange = (e) => {
    this.setState({ projectName: e.target.value });
  };

  handleCreateProject = () => {
    const { projectName, isSupervised, useWails } = this.state;
    if (!projectName) {
      alert("Please enter a project name.");
      return;
    }
    const suffix = isSupervised ? "_sup" : "_unsup";
    const prefixedProjectName = `ns_${projectName}${suffix}`;

    if (useWails && CreateProject) {
      CreateProject(prefixedProjectName)
        .then(() => {
          this.loadProjects();
          this.setState({ projectName: "" }); // Clear the input field after creation
        })
        .catch((error) => {
          console.error(
            "Failed to create project using Wails, trying Axios:",
            error
          );
          this.createProjectWithAxios(prefixedProjectName); // Fallback to Axios if Wails fails
        });
    } else {
      this.createProjectWithAxios(prefixedProjectName); // Use Axios if Wails is not available or toggle is off
    }
  };

  createProjectWithAxios = (projectName) => {
    const url = `http://${window.location.hostname}:8080/api/create-project`;
    axios
      .post(url, { projectName })
      .then(() => {
        this.loadProjects();
        this.setState({ projectName: "" }); // Clear the input field after creation
      })
      .catch((error) => {
        console.error("Failed to create project using Axios:", error);
      });
  };

  toggleUseWails = () => {
    this.setState((prevState) => ({ useWails: !prevState.useWails }));
  };

  toggleSupervised = () => {
    this.setState((prevState) => ({ isSupervised: !prevState.isSupervised }));
  };

  render() {
    const { projects, useWails, isSupervised } = this.state;

    return (
      <Box p={5}>
        <Heading as="h1" size="xl" mb={5} color="teal.600">
          OSound Projects
        </Heading>
        <VStack spacing={5} align="stretch">
          <Input
            placeholder="Enter Project Name"
            value={this.state.projectName}
            onChange={this.handleProjectNameChange}
            mb={3}
          />
          <FormControl display="flex" alignItems="center" mb={3}>
            <FormLabel htmlFor="use-wails" mb="0">
              Use Wails
            </FormLabel>
            <Switch
              id="use-wails"
              isChecked={useWails}
              onChange={this.toggleUseWails}
              colorScheme="teal"
            />
          </FormControl>
          <FormControl display="flex" alignItems="center" mb={3}>
            <FormLabel htmlFor="is-supervised" mb="0">
              Supervised
            </FormLabel>
            <Switch
              id="is-supervised"
              isChecked={isSupervised}
              onChange={this.toggleSupervised}
              colorScheme="teal"
            />
          </FormControl>
          <Button onClick={this.handleCreateProject} colorScheme="teal" mb={3}>
            Create Project
          </Button>
          <Heading as="h2" size="lg" mb={3} color="teal.600">
            Existing Projects
          </Heading>
          <List spacing={3}>
            {projects && projects.length > 0 ? (
              projects.map((project, index) => (
                <ListItem
                  key={index}
                  border="1px"
                  borderRadius="md"
                  p={2}
                  cursor="pointer"
                  _hover={{ bg: "gray.100" }}
                >
                  <Link
                    to={`/sound/${
                      project.includes("_sup") ? "supervised" : "unsupervised"
                    }/${project}`}
                  >
                    <Text fontSize="md">{project}</Text>
                  </Link>
                </ListItem>
              ))
            ) : (
              <Text fontSize="md" color="gray.500">
                No projects found.
              </Text>
            )}
          </List>
        </VStack>
      </Box>
    );
  }
}

export default OSound;
