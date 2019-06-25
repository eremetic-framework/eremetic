import React, { FormEvent, SyntheticEvent } from 'react';

import Form from 'react-bootstrap/Form';
import Col from 'react-bootstrap/Col';
import Container from 'react-bootstrap/Container';
import Button from 'react-bootstrap/Button';

import { VolumeItem, VolumeFrom } from './DynamicFormEntry';

class TaskLauncher extends React.Component {
  state = {
    validated: false,
    volumes: {},
    volumes_from: [],
  };

  handleSubmit(event: SyntheticEvent) {
    const form: HTMLFormElement = event.currentTarget as HTMLFormElement;
    if (form.checkValidity() === false) {
      event.preventDefault();
      event.stopPropagation();
    }
  }

  removeVolume(id: string) {
    let { volumes } = this.state;
    delete volumes[id];
    this.setState({ volumes: volumes });
  }

  addVolume() {
    let { volumes } = this.state;
    const key = (Math.random() + 1).toString(36).substring(7);
    volumes[key] = <VolumeItem key={key} removeFunc={() => this.removeVolume(key)} />;
    this.setState({ volumes: volumes });
  }

  removeContainerVolume(id: string) {
    let { volumes_from } = this.state;
    delete volumes_from[id];
    this.setState({ volumes_from: volumes_from });
  }

  addContainerVolume() {
    let { volumes_from } = this.state;
    const key = (Math.random() + 1).toString(36).substring(7);
    volumes_from[key] = <VolumeFrom key={key} removeFunc={() => this.removeContainerVolume(key)} />;
    this.setState({ volumes_from: volumes_from });
  }

  labels(arr, columns, width) {
    if (Object.keys(arr).length === 0) {
      return null;
    }
    return (
      <Form.Row>
        {columns.map((label: string, idx: number) => (
          <Form.Group key={idx} as={Col} md={width}>
            <Form.Label>{label}</Form.Label>
          </Form.Group>
        ))}
      </Form.Row>
    );
  }

  render() {
    const { validated, volumes, volumes_from } = this.state;
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
              <Form.Control required type="text" placeholder="alpine:3.10" />
            </Form.Group>
            <Form.Group as={Col} md="4">
              <Form.Label>Command</Form.Label>
              <Form.Control required type="text" placeholder="echo $(date)" />
            </Form.Group>
          </Form.Row>

          <Form.Row>
            <Form.Group as={Col} md="4">
              <Form.Label>CPU</Form.Label>
              <Form.Control type="number" required min="0.0" defaultValue="1.0" step="0.1" />
            </Form.Group>
            <Form.Group as={Col} md="4">
              <Form.Label>Memory (MiB)</Form.Label>
              <Form.Control type="number" required min="0.0" defaultValue="100" step="1" />
            </Form.Group>
          </Form.Row>

          <Form.Row>
            <Form.Group as={Col} md="8">
              <Form.Label>Callback URL (optional)</Form.Label>
              <Form.Control type="text" placeholder="http://localhost/callback" />
            </Form.Group>
          </Form.Row>

          <Form.Row>
            <Form.Group as={Col} md="4">
              <Form.Label>
                Volumes <Button onClick={() => this.addVolume()}>Add</Button>
              </Form.Label>
              {this.labels(volumes, ['Host', 'Container'], 5)}
              {Object.keys(volumes).map((value: string) => volumes[value])}
            </Form.Group>

            <Form.Group as={Col} md="4">
              <Form.Label>
                Volumes from Container <Button onClick={() => this.addContainerVolume()}>Add</Button>
              </Form.Label>
              {/* {Object.keys(volumes_from).length > 0 && (
                                <Form.Row>
                                    <Form.Group as={Col} md="12">
                                        <Form.Label>Container</Form.Label>
                                    </Form.Group>
                                </Form.Row>
                            )} */}
              {this.labels(volumes_from, ['Container ID'], 12)}
              {Object.keys(volumes_from).map((value: string) => volumes_from[value])}
            </Form.Group>
          </Form.Row>

          <Button type="submit">Submit form</Button>
        </Form>
      </Container>
    );
  }
}

export default TaskLauncher;
