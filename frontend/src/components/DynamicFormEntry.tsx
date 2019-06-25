import React from 'react';

import Form from 'react-bootstrap/Form';
import Col from 'react-bootstrap/Col';
import Button from 'react-bootstrap/Button';

type dynamicItemProps = {
  key: string;
  removeFunc: Function;
};

export class VolumeItem extends React.Component<dynamicItemProps> {
  render() {
    const { removeFunc, key } = this.props;
    return (
      <Form.Row key={key}>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required />
        </Form.Group>
        <Form.Group as={Col} md="5">
          <Form.Control type="text" required />
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
    const { removeFunc, key } = this.props;
    return (
      <Form.Row key={key}>
        <Form.Group as={Col} md="11">
          <Form.Control type="text" required />
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
