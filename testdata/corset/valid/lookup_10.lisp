(defun (selector) (* m1.X 2))

(module m1)
(defcolumns (X :i16) (Y :i17))
(deflookup test (Y) ((selector)))
