const bodyParser = require('body-parser');
const swaggerJsdoc = require('swagger-jsdoc');

const options = {
  definition: {
    openapi: '3.0.0',
    info: {
      title: 'Chaincode Application',
      version: '0.1.0',
      description: 'Chaincode 통신 지원 RestAPI',
    },
    // 요청 주소가 다를 경우 설정
    // servers: [{
    //     url: "https://corc.tk/chain-api",
    // }, ],
  },
  apis: ['./route/index.js'],
};

exports.specs = swaggerJsdoc(options);
