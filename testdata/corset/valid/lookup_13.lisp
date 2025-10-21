(deflookup l1 (m2.X) (m1.B))

(module m1)
(defcolumns (A :i32))
(defpermutation (B) ((+ A)))

(module m2)
(defcolumns (X :i32))
