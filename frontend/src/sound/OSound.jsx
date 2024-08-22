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
} from "@chakra-ui/react";

// Import Wails backend methods
import { CreateProject, ListProjects } from "../../wailsjs/go/main/App";

class OSound extends React.Component {
  constructor(props) {
    super(props);
    this.state = {
      projectName: "",
      selectedProject: null,
      projects: [], // List of existing projects
    };
  }

  componentDidMount() {
    this.loadProjects();
  }

  loadProjects = () => {
    ListProjects()
      .then((projects) => {
        this.setState({ projects: projects.filter((project) => project.startsWith("ns_")) });
      })
      .catch((error) => {
        console.error("Failed to load projects:", error);
      });
  };

  handleProjectNameChange = (e) => {
    this.setState({ projectName: e.target.value });
  };

  handleCreateProject = () => {
    const { projectName } = this.state;
    if (!projectName) {
      alert("Please enter a project name.");
      return;
    }
    const prefixedProjectName = `ns_${projectName}`;
    CreateProject(prefixedProjectName)
      .then(() => {
        this.loadProjects();
        this.setState({ projectName: "" }); // Clear the input field after creation
      })
      .catch((error) => {
        console.error("Failed to create project:", error);
      });
  };

  handleProjectSelection = (projectName) => {
    this.setState({ selectedProject: projectName });
    this.props.onSelectProject(projectName); // Trigger the parent to load the selected project view
  };

  render() {
    const { projects } = this.state;

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
          <Button onClick={this.handleCreateProject} colorScheme="teal" mb={3}>
            Create Project
          </Button>
          <Heading as="h2" size="lg" mb={3} color="teal.600">
            Existing Projects
          </Heading>
          <List spacing={3}>
            {projects?.map((project, index) => (
              <ListItem
                key={index}
                border="1px"
                borderRadius="md"
                p={2}
                cursor="pointer"
                onClick={() => this.handleProjectSelection(project)}
                _hover={{ bg: "gray.100" }}
              >
                <Text fontSize="md">{project}</Text>
              </ListItem>
            ))}
          </List>
        </VStack>
      </Box>
    );
  }
}

export default OSound;
