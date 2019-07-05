import React, { FormEvent, SyntheticEvent } from 'react';

import Form from 'react-bootstrap/Form';
import Col from 'react-bootstrap/Col';
import Row from 'react-bootstrap/Row';
import Container from 'react-bootstrap/Container';
import Button from 'react-bootstrap/Button';

import { VolumeItem, VolumeFrom, URI, AgentConstraint, Port, EnvironmentVariable } from './DynamicFormEntry';

class TaskLauncher extends React.Component {
  state = {
    validated: false,
    volumes: {},
    volumes_from: {},
    uris: {},
    agent_constraints: {},
    ports: {},
    envs: {},
    masked_envs: {},
  };

  handleSubmit(event: SyntheticEvent) {
    const form: HTMLFormElement = event.currentTarget as HTMLFormElement;
    if (form.checkValidity() === false) {
      event.preventDefault();
      event.stopPropagation();
    }
  }

  componentForType(type: string) {
    switch (type) {
      case 'volumes':
        return VolumeItem;
      case 'volumes_from':
        return VolumeFrom;
      case 'uris':
        return URI;
      case 'agent_constraints':
        return AgentConstraint;
      case 'ports':
        return Port;
      case 'envs':
        return EnvironmentVariable;
      case 'masked_envs':
        return EnvironmentVariable;
      default:
        console.error('unsupported type');
        return null;
    }
  }

  remove(type: string, key: string) {
    let store = this.state[type];
    delete store[key];
    this.setState({ [type]: store });
  }

  add(type: string) {
    let store = this.state[type];
    let C = this.componentForType(type);
    const key = (Math.random() + 1).toString(36).substring(7);
    store[key] = <C key={key} id={key} removeFunc={() => this.remove(type, key)} />;
    this.setState({ [type]: store });
  }

  dynamicItem({ label, collector }) {
    const store = this.state[collector];

    return (
      <Form.Group as={Col} md="4">
        <Row>
          <Col sm={9}>
            <Form.Label>{label}</Form.Label>
          </Col>
          <Col sm={2}>
            <Button
              size="sm"
              style={{ backgroundColor: '#0099C8', border: 'none' }}
              onClick={() => this.add(collector)}>
              Add
            </Button>
          </Col>
        </Row>
        {Object.keys(store).map((value: string) => store[value])}
      </Form.Group>
    );
  }

  render() {
    const { validated } = this.state;
    return (
      <Container
        style={{
          borderLeft: '5px solid black',
          borderRight: '5px solid black',
          height: '100vh',
          padding: '2em',
          backgroundColor: '#ffffff',
          fontWeight: 'bold',
        }}>
        <Form noValidate validated={validated} onSubmit={(e: FormEvent) => this.handleSubmit(e)}>
          <Form.Row>
            <Form.Group as={Col} md="4">
              <Form.Label>Docker Image</Form.Label>
              <Form.Control required type="text" placeholder="alpine:3.10" name="docker_image" />
            </Form.Group>
            <Form.Group as={Col} md="4">
              <Form.Label>Command</Form.Label>
              <Form.Control required type="text" placeholder="echo $(date)" name="command" />
            </Form.Group>
          </Form.Row>

          <Form.Row>
            <Form.Group as={Col} md="4">
              <Form.Label>CPU</Form.Label>
              <Form.Control type="number" required min="0.0" defaultValue="1.0" step="0.1" name="cpu" />
            </Form.Group>
            <Form.Group as={Col} md="4">
              <Form.Label>Memory (MiB)</Form.Label>
              <Form.Control type="number" required min="0.0" defaultValue="100" step="1" name="memory" />
            </Form.Group>
          </Form.Row>

          <Form.Row>
            <Form.Group as={Col} md="8">
              <Form.Label>Callback URL (optional)</Form.Label>
              <Form.Control type="text" placeholder="http://localhost/callback" name="callback_url" />
            </Form.Group>
          </Form.Row>

          <Form.Row>
            {this.dynamicItem({ label: 'Volumes', collector: 'volumes' })}
            {this.dynamicItem({ label: 'Volumes from Container', collector: 'volumes_from' })}
          </Form.Row>
          <Form.Row>
            {this.dynamicItem({ label: 'Environment Variables', collector: 'envs' })}
            {this.dynamicItem({ label: 'Masked Environment Variables', collector: 'masked_envs' })}
          </Form.Row>

          <Form.Row>
            {this.dynamicItem({ label: 'URIs', collector: 'uris' })}
            {this.dynamicItem({ label: 'Agent Constraints', collector: 'agent_constraints' })}
          </Form.Row>

          <Form.Row>{this.dynamicItem({ label: 'Ports', collector: 'ports' })}</Form.Row>

          <Button style={{ backgroundColor: '#0099C8', border: 'none' }} type="submit">
            Submit form
          </Button>
        </Form>
      </Container>
    );
  }
}

export default TaskLauncher;
