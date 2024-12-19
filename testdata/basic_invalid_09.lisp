;;error:5:21-25:qualified access not permitted here
(module m1)
(defcolumns X)

(defconstraint c () m1.X)
