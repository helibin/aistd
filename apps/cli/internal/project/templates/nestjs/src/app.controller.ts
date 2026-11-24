import { Controller, Get } from '@nestjs/common';

@Controller('api')
export class AppController {
  @Get('health/live') live() { return { status: 'ok' }; }
  @Get('health/ready') ready() { return { status: 'ready' }; }
  @Get('hello') hello() { return { message: 'hello' }; }
}
