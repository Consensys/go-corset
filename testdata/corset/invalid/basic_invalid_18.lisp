;;error:2:30-34:invalid integer constant
(defcolumns (X :i16 :padding 123x))
(defconstraint c1 () (!= 0 X))
