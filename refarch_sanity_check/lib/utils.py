import logging
import six


def all_nested_items(dictionary):
    """
    A generator returning key, value pairs of all items in the provided dictionary, including nested dictionaries.
    """
    for key, value in six.iteritems(dictionary):
        if isinstance(value, dict):
            # Recurse the nested dict
            for key, value in all_nested_items(value):
                yield (key, value)
        else:
            yield (key, value)


def all_nested_keys(dictionary):
    """
    A generator returning all the keys in the provided dictionary, including nested dictionaries.
    """
    for key, _ in all_nested_items(dictionary):
        yield key


def all_nested_values(dictionary):
    """
    A generator returning all the values in the provided dictionary, including nested dictionaries.
    """
    for _, value in all_nested_items(dictionary):
        yield value


def get_configured_logger():
    """
    Configures the logging settings to log more information than default and set the appropriate log level.
    """
    logger = logging.getLogger("refarch-sanity-check")
    formatter = logging.Formatter(
        fmt="%(levelname)-8s %(asctime)s    %(message)s",
        datefmt="%Y-%m-%d %H:%M:%S",
    )
    handler = logging.StreamHandler()
    handler.setFormatter(formatter)
    logger.addHandler(handler)
    return logger


def configure_loglevel(loglevel):
    # Inline import to avoid circular imports
    from . import global_vars
    global_vars.logger.setLevel(global_vars.LOG_LEVEL_MAP[loglevel])
