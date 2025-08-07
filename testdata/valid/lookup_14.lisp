(defun (m1_B) m1.B)
(deflookup l1 (m2.X) ((m1_B)))

(module m1)
(defcolumns (A :i32))
(defpermutation (B) ((+ A)))

(module m2)
(defcolumns (X :i32))
