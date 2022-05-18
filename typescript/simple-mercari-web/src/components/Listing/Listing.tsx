import React, { useState } from 'react';

const server = process.env.API_URL || 'http://127.0.0.1:9000';

interface Prop {
  onListingCompleted?: () => void;
}

type FormValue = {
  name: string;
  category: string;
  image: File | null;
};

export const Listing: React.FC<Prop> = (props) => {
  const { onListingCompleted } = props;
  const initialState: FormValue = {
    name: '',
    category: '',
    image: null,
  };
  const [values, setValues] = useState(initialState);

  const onChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    if (event.target.name === 'image') {
      setValues({ ...values, [event.target.name]: event.target.files?.item(0) ?? null });
    } else {
      setValues({ ...values, [event.target.name]: event.target.value });
    }
  };
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (values.image === null) {
      return;
    }
    const data = new FormData();
    data.append('name', values.name);
    data.append('category', values.category);
    data.append('image', values.image);

    fetch(server.concat('/items'), {
      method: 'POST',
      mode: 'cors',
      body: data,
    })
      .then((response) => response.json())
      .then((data) => {
        console.log('POST success:', data);
        onListingCompleted && onListingCompleted();
      })
      .catch((error) => {
        console.error('POST error:', error);
      });
  };
  return (
    <div className="Listing">
      <form onSubmit={onSubmit}>
        <div>
          <input
            type="text"
            name="name"
            id="name"
            placeholder="name"
            onChange={onChange}
            required
          />
          <input
            type="text"
            name="category"
            id="category"
            placeholder="category"
            onChange={onChange}
          />
          <input type="file" name="image" id="image" placeholder="image" onChange={onChange} />
          <button type="submit">List this item</button>
        </div>
      </form>
    </div>
  );
};
