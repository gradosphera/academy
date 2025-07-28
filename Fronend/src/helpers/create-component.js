import { h, render } from 'vue';

export const createComponent = (component, props, parentContainer, slots= {}, id) => {
    const vNode = h(component, props, slots);
    const container = document.createElement('div');
    container.setAttribute('id', id || '');
    parentContainer.appendChild(container);
    render(vNode, container);

    return vNode.component;
};
