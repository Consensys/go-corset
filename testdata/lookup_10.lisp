(defun (selector) (* m1.X 2))

(module m1)
(defcolumns X Y)
(deflookup test (Y) ((selector)))
