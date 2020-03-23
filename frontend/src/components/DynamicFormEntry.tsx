import React from 'react';

import Form from 'react-bootstrap/Form';
import Col from 'react-bootstrap/Col';
import Row from 'react-bootstrap/Row';
import Button from 'react-bootstrap/Button';

const componentForType = (type: string) => {
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
};

type dynamicItemProps = {
  id: string;
  removeFunc: Function;
  formOnChange: (e: React.ChangeEvent<any>) => void;
  formValues: {};
  formErrors: {};
  formTouched: {};
};
type wrapperProps = {
  label: string;
  type: string;
  formOnChange: (e: React.ChangeEvent<any>) => void;
  formValues: {};
  formErrors: {};
  formTouched: {};
};
export class DynamicItem extends React.Component<wrapperProps> {
  state = {
    entries: [],
  };

  add() {
    let { entries } = this.state;
    const key = (Math.random() + 1).toString(36).substring(7);
    entries.push(key);
    this.setState({ entries });
  }

  remove(key: string) {
    let { entries } = this.state;
    entries = entries.filter(e => e !== key);
    this.setState({ entries });
  }

  render() {
    const { label, type, formOnChange, formValues, formErrors, formTouched } = this.props;
    const { entries } = this.state;
    const C = componentForType(type);
    return (
      <Form.Group as={Col} md="4">
        <Row>
          <Col sm={9}>
            <Form.Label>{label}</Form.Label>
          </Col>
          <Col sm={2}>
            <Button size="sm" style={{ backgroundColor: '#0099C8', border: 'none' }} onClick={() => this.add()}>
              Add
            </Button>
          </Col>
        </Row>
        {entries.map((key: string) => {
          return (
            <C
              id={key}
              key={key}
              removeFunc={() => this.remove(key)}
              formOnChange={formOnChange}
              formValues={formValues}
              formErrors={formErrors}
              formTouched={formTouched}
            />
          );
        })}
      </Form.Group>
    );
  }
}

export class VolumeItem extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, id, formOnChange, formValues, formErrors, formTouched } = this.props;
    const host = `volume.${id}.host`;
    const container = `volume.${id}.container`;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control
            required
            type="text"
            placeholder="host volume"
            name={host}
            onChange={formOnChange}
            value={formValues[host]}
            isValid={formTouched[host] && !formErrors[host]}
          />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control
            required
            type="text"
            placeholder="container volume"
            name={container}
            onChange={formOnChange}
            value={formValues[container]}
            isValid={formTouched[container] && !formErrors[container]}
          />
        </Form.Group>
        <Form.Group as={Col} md="1">
          <Button variant="danger" onClick={() => removeFunc()}>
            &times;
          </Button>
        </Form.Group>
      </Form.Row>
    );
  }
}

export class EnvironmentVariable extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, id, formOnChange, formValues, formErrors, formTouched } = this.props;
    const env_key = `env.${id}.key`;
    const env_value = `env.${id}.value`;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control
            required
            type="text"
            placeholder="key"
            name={env_key}
            onChange={formOnChange}
            value={formValues[env_key]}
            isValid={formTouched[env_key] && !formErrors[env_key]}
          />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control
            required
            type="text"
            placeholder="value"
            name={env_value}
            onChange={formOnChange}
            value={formValues[env_value]}
            isValid={formTouched[env_value] && !formErrors[env_value]}
          />
        </Form.Group>
        <Form.Group as={Col} md="1">
          <Button variant="danger" onClick={() => removeFunc()}>
            &times;
          </Button>
        </Form.Group>
      </Form.Row>
    );
  }
}

export class VolumeFrom extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, id, formOnChange, formValues, formErrors, formTouched } = this.props;
    const container_key = `container_id.${id}`;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="11">
          <Form.Control
            required
            type="text"
            placeholder="container ID"
            name={container_key}
            onChange={formOnChange}
            value={formValues[container_key]}
            isValid={formTouched[container_key] && !formErrors[container_key]}
          />
        </Form.Group>
        <Form.Group as={Col} md="1">
          <Button variant="danger" onClick={() => removeFunc()}>
            &times;
          </Button>
        </Form.Group>
      </Form.Row>
    );
  }
}

export class URI extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, id, formOnChange, formValues, formErrors, formTouched } = this.props;
    const uri_key = `uri_${id}`;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="11">
          <Form.Control
            required
            type="text"
            placeholder="URI"
            name={uri_key}
            onChange={formOnChange}
            value={formValues[uri_key]}
            isValid={formTouched[uri_key] && !formErrors[uri_key]}
          />
        </Form.Group>
        <Form.Group as={Col} md="1">
          <Button variant="danger" onClick={() => removeFunc()}>
            &times;
          </Button>
        </Form.Group>
      </Form.Row>
    );
  }
}

export class AgentConstraint extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, id, formOnChange, formValues, formErrors, formTouched } = this.props;
    const constraint_name = `constraint.${id}.name`;
    const constraint_value = `constraint.${id}.value`;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control
            required
            type="text"
            placeholder="Attribute Name"
            name={constraint_name}
            onChange={formOnChange}
            value={formValues[constraint_name]}
            isValid={formTouched[constraint_name] && !formErrors[constraint_name]}
          />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control
            required
            type="text"
            placeholder="Attribute Value"
            name={constraint_value}
            onChange={formOnChange}
            value={formValues[constraint_value]}
            isValid={formTouched[constraint_value] && !formErrors[constraint_value]}
          />
        </Form.Group>
        <Form.Group as={Col} md="1">
          <Button variant="danger" onClick={() => removeFunc()}>
            &times;
          </Button>
        </Form.Group>
      </Form.Row>
    );
  }
}

export class Port extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, id, formOnChange, formValues, formErrors, formTouched } = this.props;
    const port_value = `port.${id}.value`;
    const port_type = `port.${id}.type`;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control
            required
            type="text"
            placeholder="Port"
            name={port_value}
            onChange={formOnChange}
            value={formValues[port_value]}
            isValid={formTouched[port_value] && !formErrors[port_value]}
          />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control
            as="select"
            name={port_type}
            onChange={formOnChange}
            value={formValues[port_type]}
            isValid={formTouched[port_type] && !formErrors[port_type]}>
            <option>TCP</option>
            <option>UDP</option>
          </Form.Control>
        </Form.Group>
        <Form.Group as={Col} md="1">
          <Button variant="danger" onClick={() => removeFunc()}>
            &times;
          </Button>
        </Form.Group>
      </Form.Row>
    );
  }
}
