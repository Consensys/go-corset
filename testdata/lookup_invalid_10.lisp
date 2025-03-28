;;error:2:15-19:conflicting context
(defun (test) m1.Z)

(module m1)
(defcolumns (X :i16) (Y :i16) (Z :i16))

(module m2)
(defcolumns (A :i16))

(deflookup l1 (m1.X m1.Y) (A (test)))
