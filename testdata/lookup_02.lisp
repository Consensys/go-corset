(defcolumns (A :i16) (B :i8) (X :i16) (Y :i8))
;; Inclusion A,B into X,Y
(deflookup test (X Y) (A B))
