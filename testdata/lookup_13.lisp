(deflookup l1 (m1.B) (m2.X))

(module m1)
(defcolumns (A :i32))
(defpermutation (B) ((+ A)))

(module m2)
(defcolumns (X :i32))
