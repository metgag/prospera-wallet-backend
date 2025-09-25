-- Active: 1758768363215@@127.0.0.1@5434@prospera@public
INSERT INTO accounts (email, password, pin)
VALUES
('user1@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '123456'),
('user2@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '234567'),
('user3@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '345678'),
('user4@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '456789'),
('user5@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '567890'),
('user6@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '678901'),
('user7@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '789012'),
('user8@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '890123'),
('user9@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '901234'),
('user10@mail.com', '$argon2id$v=19$m=65536,t=2,p=1$nPyzjKxUHSuHJlem+f3MrQ$G4jFgcrXOvnXHdstUhTslhAQUkoMKmKT71SIRT6newo', '012345');

INSERT INTO profiles (fullname, phone, img, verified)
VALUES
('Rangga Putra', '0811111111', 'rangga.png', TRUE),
('Budi Santoso', '0822222222', 'budi.png', FALSE),
('Siti Aminah', '0833333333', 'siti.png', TRUE),
('Andi Wijaya', '0844444444', 'andi.png', FALSE),
('Dewi Lestari', '0855555555', 'dewi.png', TRUE),
('Agus Pratama', '0866666666', 'agus.png', FALSE),
('Tono Saputra', '0877777777', 'tono.png', TRUE),
('Joko Susilo', '0888888888', 'joko.png', FALSE),
('Lina Marlina', '0899999999', 'lina.png', TRUE),
( 'Wati Kurnia', '0800000000', 'wati.png', FALSE);

INSERT INTO wallets (balance)
VALUES
(1000000),
(500000),
(250000),
(750000),
(1200000),
(900000),
(300000),
(100000),
(450000),
(800000);

INSERT INTO internal_accounts (name, img, tax)
VALUES
('TopUp Service', 'topup.png', 1000),
('Admin Fee', 'admin.png', 2000),
('Tax Office', 'tax.png', 1500),
('System Reserve', 'reserve.png', 500),
('Promo Cashback', 'promo.png', 0);
