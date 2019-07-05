import React from 'react';

import Form from 'react-bootstrap/Form';
import Col from 'react-bootstrap/Col';
import Button from 'react-bootstrap/Button';

type dynamicItemProps = {
  id: string;
  removeFunc: Function;
};

export class VolumeItem extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, id } = this.props;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required placeholder="host volume" name={`volume[${id}][host]`} />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required placeholder="container volume" name={`volume[${id}][container]`} />
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
    const { removeFunc, id } = this.props;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required placeholder="key" name={`env_key_${id}`} />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required placeholder="value" name={`env_value_${id}`} />
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
    const { removeFunc, id } = this.props;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="11">
          <Form.Control type="text" required placeholder="container ID" name={`container_id_${id}`} />
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
    const { removeFunc, id } = this.props;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="11">
          <Form.Control type="text" required placeholder="URI" name={`uri_${id}`} />
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
    const { removeFunc, id } = this.props;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required placeholder="Attribute Name" name={`constraint_name_${id}`} />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required placeholder="Attribute Value" name={`constraint_value_${id}`} />
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
    const { removeFunc, id } = this.props;
    return (
      <Form.Row key={id}>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required placeholder="Port" name={`port_value_${id}`} />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control as="select" name={`port_type_${id}`}>
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
