(defcolumns (X :u16))
(defproperty lem (if
                  (!= 0 X)
                  (!= X (shift X -1))))
