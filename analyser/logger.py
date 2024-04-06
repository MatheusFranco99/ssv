import logging

class Logger:
    def __init__(self, output_file, verbose_mode=1):
        self.verbose_mode = verbose_mode
        self.logger = logging.getLogger(__name__)
        self.logger.setLevel(logging.DEBUG)
        
        formatter = logging.Formatter('%(levelname)s -  %(message)s')
        
        console_handler = logging.StreamHandler()
        console_handler.setLevel(logging.DEBUG)
        console_handler.setFormatter(formatter)
        self.logger.addHandler(console_handler)
        
        file_handler = logging.FileHandler(output_file)
        file_handler.setLevel(logging.DEBUG)
        file_handler.setFormatter(formatter)
        self.logger.addHandler(file_handler)
    
        self.tag = 0
    
    def newLogger(self, prefix):
        new_logger = logging.LoggerAdapter(self.logger, {'prefix': prefix})
        return new_logger
    
    def debug(self, message, *args, **kwargs):
        if self.verbose_mode == 1:
            self.logger.debug(message, *args, **kwargs)
    
    def info(self, message, *args, **kwargs):
        if self.verbose_mode != 0:
            self.logger.info(message, *args, **kwargs)

    def warning(self,message):
        self.logger.warning(message)
    
    def getTag(self):
        self.tag += 1
        return self.tag