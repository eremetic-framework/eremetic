import React, { FormEvent, SyntheticEvent } from 'react';
import { Formik } from 'formik';
import Form from 'react-bootstrap/Form';
import Col from 'react-bootstrap/Col';
import Row from 'react-bootstrap/Row';
import Container from 'react-bootstrap/Container';
import Button from 'react-bootstrap/Button';

import {
  VolumeItem,
  VolumeFrom,
  URI,
  AgentConstraint,
  Port,
  EnvironmentVariable,
  DynamicItem,
} from './DynamicFormEntry';

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

  render() {
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
        <Formik
          onSubmit={values => {
            console.log(values);
          }}
          initialValues={{
            docker_image: '',
            command: '',
            cpu: 1.0,
            memory: 100,
            callback_url: '',
            volume: {},
            env: {},
            container_id: {},
            uri: {},
            constraints: {},
            port: {},
          }}>
          {({ handleSubmit, handleChange, values, touched, errors }) => (
            <Form noValidate onSubmit={handleSubmit}>
              <Form.Row>
                <Form.Group as={Col} md="4">
                  <Form.Label>Docker Image</Form.Label>
                  <Form.Control
                    required
                    type="text"
                    placeholder="alpine:3.10"
                    name="docker_image"
                    onChange={handleChange}
                    value={values['docker_image']}
                    isValid={touched['docker_image'] && !errors['docker_image']}
                  />
                  <Form.Control.Feedback type="invalid">{errors['docker_image']}</Form.Control.Feedback>
                </Form.Group>
                <Form.Group as={Col} md="4">
                  <Form.Label>Command</Form.Label>
                  <Form.Control
                    required
                    type="text"
                    placeholder="echo $(date)"
                    name="command"
                    onChange={handleChange}
                    value={values.command}
                    isValid={touched.command && !errors.command}
                  />
                </Form.Group>
              </Form.Row>

              <Form.Row>
                <Form.Group as={Col} md="4">
                  <Form.Label>CPU</Form.Label>
                  <Form.Control
                    type="number"
                    required
                    min="0.0"
                    step="0.1"
                    name="cpu"
                    onChange={handleChange}
                    value={values.cpu}
                    isValid={touched.cpu && !errors.cpu}
                  />
                </Form.Group>
                <Form.Group as={Col} md="4">
                  <Form.Label>Memory (MiB)</Form.Label>
                  <Form.Control
                    type="number"
                    required
                    min="0.0"
                    step="1"
                    name="memory"
                    onChange={handleChange}
                    value={values.memory}
                    isValid={touched.memory && !errors.memory}
                  />
                </Form.Group>
              </Form.Row>

              <Form.Row>
                <Form.Group as={Col} md="8">
                  <Form.Label>Callback URL (optional)</Form.Label>
                  <Form.Control
                    type="text"
                    placeholder="http://localhost/callback"
                    name="callback_url"
                    onChange={handleChange}
                    value={values.callback_url}
                    isValid={touched.callback_url && !errors.callback_url}
                  />
                </Form.Group>
              </Form.Row>

              <Form.Row>
                <DynamicItem
                  label="Volumes"
                  type={'volumes'}
                  formOnChange={handleChange}
                  formValues={values}
                  formTouched={touched}
                  formErrors={errors}
                />
                <DynamicItem
                  label="Volumes from Container"
                  type={'volumes_from'}
                  formOnChange={handleChange}
                  formValues={values}
                  formTouched={touched}
                  formErrors={errors}
                />
              </Form.Row>
              <Form.Row>
                <DynamicItem
                  label="Environment Variables"
                  type={'envs'}
                  formOnChange={handleChange}
                  formValues={values}
                  formTouched={touched}
                  formErrors={errors}
                />
                <DynamicItem
                  label="Masked Environment Variables"
                  type={'masked_envs'}
                  formOnChange={handleChange}
                  formValues={values}
                  formTouched={touched}
                  formErrors={errors}
                />
              </Form.Row>

              <Form.Row>
                <DynamicItem
                  label="URIs"
                  type={'uris'}
                  formOnChange={handleChange}
                  formValues={values}
                  formTouched={touched}
                  formErrors={errors}
                />
                <DynamicItem
                  label="Agent Constraints"
                  type={'agent_constraints'}
                  formOnChange={handleChange}
                  formValues={values}
                  formTouched={touched}
                  formErrors={errors}
                />
              </Form.Row>

              <Form.Row>
                <DynamicItem
                  label="Ports"
                  type={'ports'}
                  formOnChange={handleChange}
                  formValues={values}
                  formTouched={touched}
                  formErrors={errors}
                />
              </Form.Row>

              <Button style={{ backgroundColor: '#0099C8', border: 'none' }} type="submit">
                Submit form
              </Button>
            </Form>
          )}
        </Formik>
      </Container>
    );
  }
}

export default TaskLauncher;
