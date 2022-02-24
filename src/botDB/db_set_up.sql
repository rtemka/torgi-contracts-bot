
CREATE DATABASE purchase_registry WITH ENCODING 'UTF8' LC_COLLATE='ru_RU.UTF-8' LC_CTYPE='ru_RU.UTF-8' TEMPLATE=template0;

CREATE TABLE IF NOT EXISTS customer_types (
	customer_type_id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	customer_type_name varchar(10) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS purchase_types (
	purchase_type_id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	purchase_type_name varchar(10) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS regions (
	region_id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	region_name varchar(50) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS etp (
	etp_id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	etp_name varchar(20) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS statuses (
	status_id integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	status_name varchar(20) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS purchase_string_codes (
	purchase_string_code integer GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	purchase_string_code_name varchar(5) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS purchase_registry (
	registry_number varchar (20) PRIMARY KEY,
	purchase_id bigint GENERATED ALWAYS AS IDENTITY,
	purchase_subject text NOT NULL,
	purchase_string_code integer NOT NULL,
	purchase_type_id integer NOT NULL,
	-- collecting date NOT NULL,
	-- collecting_time timetz NOT NULL,
	collecting timestamptz NOT NULL,
	approval date,
	-- bidding_date date,
	-- bidding_time timetz,
	bidding timestamptz,
	region_id integer NOT NULL,
	customer_type_id integer NOT NULL,
	max_price numeric(16, 2) NOT NULL,
	application_guarantee numeric(16, 2),
	contract_guarantee numeric(16, 2),
	status_id integer,
	our_participants varchar(100),
	estimation numeric(8, 2),
	etp_id integer,
	winner varchar(300),
	winner_price numeric(16, 2),
	participants varchar(600),
	FOREIGN KEY (purchase_type_id) REFERENCES purchase_types (purchase_type_id),
	FOREIGN KEY (region_id) REFERENCES regions (region_id),
	FOREIGN KEY (customer_type_id) REFERENCES customer_types (customer_type_id),
	FOREIGN KEY (etp_id) REFERENCES etp (etp_id),
	FOREIGN KEY (status_id) REFERENCES statuses (status_id),
	FOREIGN KEY (purchase_string_code) REFERENCES purchase_string_codes (purchase_string_code)
);