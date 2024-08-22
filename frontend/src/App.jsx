import { useState } from "react";
import logo from "./assets/images/logo-universal.png";
import "./App.css";
import { Greet } from "../wailsjs/go/main/App";

import { ChakraProvider } from "@chakra-ui/react";
import OHome from "./OHome";

function OLDApp() {
  const [resultText, setResultText] = useState(
    "Please enter your name below please 👇"
  );
  const [name, setName] = useState("");
  const updateName = (e) => setName(e.target.value);
  const updateResultText = (result) => setResultText(result);

  function greet() {
    Greet(name).then(updateResultText);
  }

  return (
    <div id="App">
      <img src={logo} id="logo" alt="logo" />
      <h1>Welcome</h1>
      <div id="result" className="result">
        {resultText}
      </div>
      <div id="input" className="input-box">
        <input
          id="name"
          className="input"
          onChange={updateName}
          autoComplete="off"
          name="input"
          type="text"
        />
        <button className="btn" onClick={greet}>
          Greet
        </button>
      </div>
    </div>
  );
}

function App() {
  return (
    <ChakraProvider>
      <OHome />
    </ChakraProvider>
  );
}

export default App;
