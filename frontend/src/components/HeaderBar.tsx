import React from 'react';
import Head from 'next/head';

import Navbar from 'react-bootstrap/Navbar';
import Nav from 'react-bootstrap/Nav';
import Form from 'react-bootstrap/Form';
import FormControl from 'react-bootstrap/FormControl';
import Button from 'react-bootstrap/Button';

class HeaderBar extends React.Component {
  randomBackground() {
    const c = Math.floor(Math.random() * Math.floor(7)) + 1;
    return `/static/crabs/hermit_00${c}.jpg`;
  }

  render() {
    return (
      <React.Fragment>
        <Head>
          <link
            rel="stylesheet"
            href="https://maxcdn.bootstrapcdn.com/bootstrap/4.3.1/css/bootstrap.min.css"
            integrity="sha384-ggOyR0iXCbMQv3Xipma34MD+dH/1fQ784/j6cY/iJTQUOhcWr7x9JvoRxT2MZw1T"
            crossOrigin="anonymous"
          />
        </Head>
        <style jsx global>
          {`
                    body {
                        background-image: url("${this.randomBackground()}");
                        background-size: cover;
                        background-repeat: no-repeat;
                        background-attachment: fixed;
                        font: 11px menlo;
                    }`}
        </style>
        <Navbar style={{ backgroundColor: '#d3d3ee', borderBottom: '5px solid black' }}>
          <Navbar.Brand href="/tasks">
            <img alt="eremetic_logo" src="/static/eremetic_logo.png" height="60" />
          </Navbar.Brand>
          <Nav className="mr-auto">
            <Nav.Link
              href="/launch"
              style={{
                color: 'black',
                fontWeight: 'bold',
              }}>
              <Button style={{ backgroundColor: '#0099C8', border: 'none' }}>Launch Task</Button>
            </Nav.Link>
          </Nav>
          <Form inline>
            <FormControl style={{ width: '40em' }} type="text" placeholder="Task ID" />
            <Button style={{ backgroundColor: '#0099C8', border: 'none' }}>View</Button>
          </Form>
        </Navbar>
      </React.Fragment>
    );
  }
}

export default HeaderBar;
